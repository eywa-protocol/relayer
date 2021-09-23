package test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol/modelBLS"
	messageSigpb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol/protobuf/messageWithSig"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol/protobuf/messagepb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol/test_utils"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"

	core "github.com/libp2p/go-libp2p-core"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/protocol/model"
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
		var comm *protocol.Libp2pPubSub
		comm = new(protocol.Libp2pPubSub)
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
			fmt.Printf("LOCAL HOST %s CONNECT TO nxtHost %s \n", protocol.GetLocalhostAddress(*hosts[i]), protocol.GetLocalhostAddress(*nxtHost))

			err := protocol.ConnectHostToPeerWithError(*hosts[i], protocol.GetLocalhostAddress(*nxtHost))
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i < n; i++ {
		err = protocol.ConnectHostToPeerWithError(*hosts[i], protocol.GetLocalhostAddress(*hosts[(i+1)%n]))
		if err != nil {
			return
		}
		err = protocol.ConnectHostToPeerWithError(*hosts[i], protocol.GetLocalhostAddress(*hosts[(i+2)%n]))
		if err != nil {
			return
		}
		err = protocol.ConnectHostToPeerWithError(*hosts[i], protocol.GetLocalhostAddress(*hosts[(i+3)%n]))
		if err != nil {
			return
		}
		err = protocol.ConnectHostToPeerWithError(*hosts[i], protocol.GetLocalhostAddress(*hosts[(i+4)%n]))
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
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	err := setupNetworkTopology(hosts)
	require.Nil(t, err)
	// PubSub is ready and we can start our algorithm
	StartTestBLS(nodes, stop, stop/3)
	LogOutputBLS(t, nodes)
}

func simpleTestOneStepBLS(t *testing.T, n int, initialPort int) {
	nodes, hosts := setupHostsBLS(n, initialPort)

	defer func() {
		fmt.Println("Closing hosts")
		for _, h := range hosts {
			_ = (*h).Close()
		}
	}()

	err := setupNetworkTopology(hosts)
	require.Nil(t, err)
	// PubSub is ready and we can start our algorithm
	cch := StartTestOneStepBLS(nodes)
	for _, c := range cch {
		require.True(t, c)
	}
	LogOutputBLS(t, nodes)
}

func setupHostsBLS(n int, initialPort int) ([]*modelBLS.Node, []*core.Host) {
	// nodes used in tlc model
	nodes := make([]*modelBLS.Node, n)

	// hosts used in libp2p communications
	hosts := make([]*core.Host, n)

	publicKeys := make([]kyber.Point, 0)
	privateKeys := make([]kyber.Scalar, 0)
	suite := pairing.NewSuiteBn256()

	for range nodes {
		prvKey := suite.Scalar().Pick(suite.RandomStream())
		privateKeys = append(privateKeys, prvKey)
		// list of public keys
		publicKeys = append(publicKeys, suite.Point().Mul(prvKey, nil))
	}

	fmt.Println(publicKeys)

	for i := range nodes {

		// var comm modelBLS.CommunicationInterface
		var comm *protocol.Libp2pPubSub
		comm = new(protocol.Libp2pPubSub)
		comm.SetTopic("TLC")

		// creating libp2p hosts
		host := comm.CreatePeer(i, initialPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSub(*host)
		comm.InitializeVictim(false)
		mask, _ := sign.NewMask(suite, publicKeys, nil)
		// ////

		nodes[i] = &modelBLS.Node{
			Id:           i,
			TimeStep:     0,
			ThresholdWit: n/2 + 1,
			ThresholdAck: n/2 + 1,
			Acks:         0,
			ConvertMsg:   &messageSigpb.Convert{},
			Comm:         comm,
			History:      make([]modelBLS.MessageWithSig, 0),
			Signatures:   make([][]byte, n),
			SigMask:      mask,
			PublicKeys:   publicKeys,
			PrivateKey:   privateKeys[i],
			Suite:        suite,
		}

	}
	return nodes, hosts
}

// StartTest is used for starting tlc nodes
func StartTestBLS(nodes []*modelBLS.Node, stop int, fails int) {
	fmt.Print("START")
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		node.Advance(0)
	}
	for _, node := range nodes {
		wg.Add(1)
		go runNodeBLS(node, stop, wg)
	}
	wg.Add(-fails)
	wg.Wait()
	fmt.Println("The END")
}

// StartTest is used for starting tlc nodes
func StartTestOneStepBLS(nodes []*modelBLS.Node) (consensuses []bool) {
	fmt.Print("START")
	wg := &sync.WaitGroup{}
	defer wg.Done()
	for _, node := range nodes {

		node.Advance(0)
	}
	var consensusesChan []chan bool
	for _, node := range nodes {
		wg.Add(1)
		consensusChannel := make(chan bool)
		go runOneStepNodeBLS(node, wg, consensusChannel)
		consensusesChan = append(consensusesChan, consensusChannel)

	}

	for i := 0; i < len(consensusesChan); i++ {
		consensuses = append(consensuses, <-consensusesChan[i])
	}
	fmt.Println("CONSENSUSES", consensuses)

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

func runOneStepNodeBLS(node *modelBLS.Node, wg *sync.WaitGroup, consensusChannel chan bool) {
	defer wg.Done()
	//consensusChannel := make(chan bool)
	wg.Add(1)
	go node.WaitForProtocolMsg(consensusChannel, wg)

	//achieved = <-consensusChannel
	//return
}
