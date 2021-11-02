package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/eywa-protocol/bls-crypto/bls"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/model"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/modelBLS"
	messageSigpb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/protobuf/messageWithSig"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/protobuf/messagepb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/test_utils"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
)

type FailureModel int

const (
	NoFailure = iota
	MinorFailure
	MajorFailure
	RejoiningMinorityFailure
	RejoiningMajorityFailure
	LeaveRejoin
	ThreeGroups
)

const FailureDelay = 3
const RejoinDelay = 15
const LeaveDelay = 10

// setupHosts is responsible for creating tlc nodes and also libp2p hosts.
func setupHosts(n int, initialPort int, failureModel FailureModel) ([]*model.Node, []*core.Host) {
	fmt.Print("SETUP HOSTS\n")
	// nodes used in tlc model
	nodes := make([]*model.Node, n)

	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)

	for i := range nodes {

		// var comm modelBLS.CommunicationInterface
		var comm *consensus.Protocol
		comm = new(consensus.Protocol)
		comm.SetTopic("TLC")

		// creating libp2p hosts
		host := comm.CreatePeer(i, initialPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSub(*host)

		// Simulating rejoining failures, where a node leaves the delayed set and joins other progressing nodes
		nVictim := 0
		switch failureModel {
		case RejoiningMinorityFailure:
			nVictim = (n - 1) / 2
		case RejoiningMajorityFailure:
			nVictim = (n + 1) / 2
		case LeaveRejoin:
			nVictim = (n - 1) / 2
		case ThreeGroups:
			nVictim = n
		}
		if failureModel == ThreeGroups {
			if i < 3 {
				comm.JoinGroup([]int{0, 1, 2})
			} else if i < 6 {
				comm.JoinGroup([]int{3, 4, 5})
			} else {
				comm.JoinGroup([]int{})
			}
		}

		if i < nVictim {
			comm.InitializeVictim(true)

			go func() {
				time.Sleep(2 * FailureDelay * time.Second)
				comm.AttackVictim()
			}()

			nRejoin := 2
			if failureModel == ThreeGroups {
				nRejoin = 6
			}
			if i < nRejoin {
				go func() {
					// Delay for the node to get out of delayed(victim) group
					time.Sleep((RejoinDelay + time.Duration(FailureDelay*i)) * time.Second)

					comm.ReleaseVictim()
				}()
			}
		} else {
			if failureModel == LeaveRejoin {
				if i == (n - 1) {
					go func() {
						// Delay for the node to leave the progressing group
						time.Sleep(LeaveDelay * time.Second)
						comm.Disconnect()
					}()
				}
			}

			comm.InitializeVictim(false)
		}

		nodes[i] = &model.Node{
			Id:           i,
			TimeStep:     0,
			ThresholdWit: n/2 + 1,
			ThresholdAck: n/2 + 1,
			Acks:         0,
			ConvertMsg:   &messagepb.Convert{},
			Comm:         comm,
			History:      make([]model.Message, 0)}
	}
	return nodes, hosts
}

// setupNetworkTopology sets up a simple network topology for test.
func setupNetworkTopology(hosts []*core.Host) (err error) {

	// Connect hosts to each other in a topology
	n := len(hosts)

	for i := 0; i < n; i++ {
		for j, nxtHost := range hosts {
			if j == i {
				continue
			}
			fmt.Printf("LOCAL HOST %s CONNECT TO nxtHost %s \n", consensus.GetLocalhostAddress(*hosts[i]), consensus.GetLocalhostAddress(*nxtHost))

			err := consensus.ConnectHostToPeerWithError(*hosts[i], consensus.GetLocalhostAddress(*nxtHost))
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i < n; i++ {
		err = consensus.ConnectHostToPeerWithError(*hosts[i], consensus.GetLocalhostAddress(*hosts[(i+1)%n]))
		if err != nil {
			return
		}
		err = consensus.ConnectHostToPeerWithError(*hosts[i], consensus.GetLocalhostAddress(*hosts[(i+2)%n]))
		if err != nil {
			return
		}
		err = consensus.ConnectHostToPeerWithError(*hosts[i], consensus.GetLocalhostAddress(*hosts[(i+3)%n]))
		if err != nil {
			return
		}
		err = consensus.ConnectHostToPeerWithError(*hosts[i], consensus.GetLocalhostAddress(*hosts[(i+4)%n]))
		if err != nil {
			return
		}
	}
	// Wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)
	return err
}

func minorityFailure(nodes []*model.Node, n int) int {
	nFail := (n - 1) / 2
	// nFail := 4
	go func(nodes []*model.Node, nFail int) {
		time.Sleep(FailureDelay * time.Second)
		failures(nodes, nFail)
	}(nodes, nFail)

	return nFail
}

func majorityFailure(nodes []*model.Node, n int) int {
	nFail := n/2 + 1
	go func(nodes []*model.Node, nFail int) {
		time.Sleep(FailureDelay * time.Second)
		failures(nodes, nFail)
	}(nodes, nFail)
	return nFail
}

func failures(nodes []*model.Node, nFail int) {
	for i, node := range nodes {
		if i < nFail {
			node.Comm.Disconnect()
		}
	}
}

func simpleTest(t *testing.T, n int, initialPort int, stop int, failureModel FailureModel) {
	var nFail int
	nodes, hosts := setupHosts(n, initialPort, failureModel)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	err := setupNetworkTopology(hosts)
	require.Nil(t, err)

	// Put failures here
	switch failureModel {
	case MinorFailure:
		nFail = minorityFailure(nodes, n)
	case MajorFailure:
		nFail = majorityFailure(nodes, n)
	case RejoiningMinorityFailure:
		nFail = (n-1)/2 - 1
	case RejoiningMajorityFailure:
		nFail = (n+1)/2 - 1
	case LeaveRejoin:
		nFail = (n-1)/2 - 1
	case ThreeGroups:
		nFail = (n - 1) / 2
	}

	// PubSub is ready and we can start our algorithm
	fmt.Printf("STARTING TEST for %d nodes stop %d nfail %d\n", len(nodes), stop, nFail)

	test_utils.StartTest(nodes, stop, nFail)
	test_utils.LogOutput(t, nodes)
}

// Testing TLC with majority thresholds with no node failures
func TestWithNoFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/TestWithNoFailure.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	// Delayed = false
	simpleTest(t, 9, 9900, 9, NoFailure)
}

// Testing TLC with minor nodes failing
func TestWithMinorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/TestWithMinorFailure.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 5, MinorFailure)
}

// Testing TLC with majority of nodes failing
func TestWithMajorFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/log4.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 10, 9900, 5, MajorFailure)
}

// Testing TLC with majority of nodes working correctly and a set of delayed nodes. a node will leave the victim set
// after some seconds and rejoin to the progressing nodes.
func TestWithRejoiningMinorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/RejoiningMinority.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 7, 9900, 3, RejoiningMinorityFailure)
}

// Testing TLC with majority of nodes being delayed. a node will leave the victim set after some seconds and rejoin to
// the other connected nodes. This will make progress possible.
func TestWithRejoiningMajorityFailure(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/RejoiningMajority.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 10, RejoiningMajorityFailure)
}

// Testing TLC with majority of nodes working correctly and a set of delayed nodes. a node will lose connection to
// progressing nodes and will stop the progress. After some seconds another node will join to the set, making progress
// possible.
func TestWithLeaveRejoin(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/log8.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 8, LeaveRejoin)
}

// TODO: Find a way to simualte this onw, since I have removed the case for this simulation
func TestWithThreeGroups(t *testing.T) {
	// Create hosts in libp2p
	logFile, _ := os.OpenFile("../../../logs/TestWithThreeGroups.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTest(t, 11, 9900, 8, ThreeGroups)
}

func TestBLS(t *testing.T) {
	logFile, _ := os.OpenFile("../../../logBLS.log", os.O_RDWR|os.O_CREATE, 0666)
	modelBLS.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	// Delayed = false
	simpleTestBLS(t, 5, 9900, 2)
}

func TestOneStepBLS(t *testing.T) {
	logFile, _ := os.OpenFile("../logs/oneStepBLS.log", os.O_RDWR|os.O_CREATE, 0666)
	modelBLS.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTestOneStepBLS(t, 7, 9900)
}

func simpleTestBLS(t *testing.T, n int, initialPort int, stop int) {
	nodes, hosts := setupHostsBLS(n, initialPort)

	defer func() {
		logrus.Info("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	err := setupNetworkTopology(hosts)
	require.Nil(t, err)
	// PubSub is ready and we can start our algorithm
	sessions := StartTestBLS(nodes, stop, stop/3)
	LogOutputBLS(t, sessions)
}

func simpleTestOneStepBLS(t *testing.T, n int, initialPort int) {
	nodes, hosts := setupHostsBLS(n, initialPort)

	defer func() {
		logrus.Info("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	err := setupNetworkTopology(hosts)
	require.Nil(t, err)
	// PubSub is ready and we can start our algorithm
	cch, sessions := StartTestOneStepBLS(nodes)
	for _, c := range cch {
		require.True(t, c)
	}
	LogOutputBLS(t, sessions)
}

func setupHostsBLS(n int, initialPort int) ([]*bridge.Node, []*core.Host) {
	// nodes used in tlc model
	nodes := make([]*bridge.Node, n)

	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)

	publicKeys := make([]bls.PublicKey, 0)
	privateKeys := make([]bls.PrivateKey, 0)

	for range nodes {
		priv, pub := bls.GenerateRandomKey()
		privateKeys = append(privateKeys, priv)
		publicKeys = append(publicKeys, pub)
	}

	anticoefs := bls.CalculateAntiRogueCoefficients(publicKeys)
	aggregatedPublicKey := bls.AggregatePublicKeys(publicKeys, anticoefs)
	aggregatedPublicKey.Marshal() // to avoid data race because Marshal() wants to normalize the key for the first time

	for i := range nodes {
		// var comm modelBLS.CommunicationInterface
		var comm *consensus.Protocol
		comm = new(consensus.Protocol)
		comm.SetTopic("TLC")

		// creating libp2p hosts
		host := comm.CreatePeer(i, initialPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSub(*host)
		comm.InitializeVictim(false)

		nodes[i] = &bridge.Node{
			Node: base.Node{
				Ctx:  context.Background(),
				Host: *host,
			},
			EpochKeys: bridge.EpochKeys{
				Id:             i,
				PublicKeys:     publicKeys,
				EpochPublicKey: aggregatedPublicKey,
			},
			PrivKey:   privateKeys[i],
			P2PPubSub: comm,
		}
	}
	return nodes, hosts
}

func makeSession(n int, node *bridge.Node) *modelBLS.Node {
	return &modelBLS.Node{
		Ctx:            node.Ctx,
		Id:             node.Id,
		TimeStep:       0,
		ThresholdWit:   n/2 + 1,
		ThresholdAck:   n/2 + 1,
		Acks:           0,
		ConvertMsg:     &messageSigpb.Convert{},
		Comm:           node.P2PPubSub,
		History:        make([]modelBLS.MessageWithSig, 0),
		SigMask:        bls.EmptyMultisigMask(),
		PublicKeys:     node.PublicKeys,
		PrivateKey:     node.PrivKey,
		EpochPublicKey: node.EpochPublicKey,
		MembershipKey:  node.MembershipKey,
		PartPublicKey:  bls.ZeroPublicKey(),
		PartSignature:  bls.ZeroSignature(),
	}
}

func StartBlsSetup(nodes []*bridge.Node) {
	wg := &sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go runNodeBLSSetup(node, wg)
	}
	msg := bridge.MessageBlsSetup{MsgType: bridge.BlsSetupPhase}
	msgBytes, _ := json.Marshal(msg)
	nodes[0].P2PPubSub.Broadcast(msgBytes)
	wg.Wait()
	logrus.Info("BLS setup done")
}

// StartTest is used for starting tlc nodes
func StartTestBLS(nodes []*bridge.Node, stop int, fails int) (sessions []*modelBLS.Node) {
	logrus.Info("START")
	n := len(nodes)

	StartBlsSetup(nodes)

	wg := &sync.WaitGroup{}
	sessions = make([]*modelBLS.Node, n)
	for i, node := range nodes {
		sessions[i] = makeSession(n, node)
	}
	sessions[0].Advance(0)

	for _, session := range sessions {
		wg.Add(1)
		go runNodeBLS(session, stop, wg)
	}

	wg.Add(-fails)
	wg.Wait()
	fmt.Println("The END")
	return
}

// StartTest is used for starting tlc nodes
func StartTestOneStepBLS(nodes []*bridge.Node) (consensuses []bool, sessions []*modelBLS.Node) {
	logrus.Info("START")
	n := len(nodes)

	StartBlsSetup(nodes)

	sessions = make([]*modelBLS.Node, n)
	for i, node := range nodes {
		sessions[i] = makeSession(n, node)
	}
	sessions[0].Advance(0)

	wg := &sync.WaitGroup{}
	var consensusesChan []chan bool
	for _, session := range sessions {
		wg.Add(1)
		consensusChannel := make(chan bool)
		go session.WaitForProtocolMsg(consensusChannel, wg)
		consensusesChan = append(consensusesChan, consensusChannel)
	}

	for i := 0; i < len(consensusesChan); i++ {
		consensuses = append(consensuses, <-consensusesChan[i])
	}
	fmt.Println("CONSENSUSES", consensuses)
	wg.Wait()
	return
}

func LogOutputBLS(t *testing.T, nodes []*modelBLS.Node) {
	for i := range nodes {
		t.Logf("LogOut nodes: %d , TimeStep : %d", i, nodes[i].TimeStep)
		modelBLS.Logger1.Printf("%d,%d\n", i, nodes[i].TimeStep)
	}
}

func runNodeBLS(node *modelBLS.Node, stop int, wg *sync.WaitGroup) {
	defer wg.Done()
	err := node.WaitForMsg(stop)
	if err != nil {
		fmt.Println(err)
	}

}

func runNodeBLSSetup(node *bridge.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	node.MembershipKey = node.BlsSetup(wg)
}