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
	messageSigpb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/protobuf/message"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
)

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

func TestOneStepBLS(t *testing.T) {
	logFile, _ := os.OpenFile("../logs/oneStepBLS.log", os.O_RDWR|os.O_CREATE, 0666)
	model.Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	simpleTestOneStepBLS(t, 7, 9900)
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

		// creating libp2p hosts
		host := comm.CreatePeer(i, initialPort+i)
		hosts[i] = host

		// creating pubsubs
		comm.InitializePubSubWithTopic(*host, "TEST")

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

func makeSession(n int, node *bridge.Node) *model.Node {
	return &model.Node{
		Ctx:            node.Ctx,
		Id:             node.Id,
		TimeStep:       0,
		ThresholdWit:   n/2 + 1,
		ThresholdAck:   n/2 + 1,
		Acks:           0,
		ConvertMsg:     &messageSigpb.Convert{},
		Comm:           node.P2PPubSub,
		History:        make([]model.MessageWithSig, 0),
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
func StartTestOneStepBLS(nodes []*bridge.Node) (consensuses []bool, sessions []*model.Node) {
	logrus.Info("START")
	n := len(nodes)

	StartBlsSetup(nodes)

	sessions = make([]*model.Node, n)
	for i, node := range nodes {
		sessions[i] = makeSession(n, node)
	}
	wg := &sync.WaitGroup{}
	sessions[0].AdvanceStep(0)

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

func LogOutputBLS(t *testing.T, nodes []*model.Node) {
	for i := range nodes {
		t.Logf("LogOut nodes: %d , TimeStep : %d", i, nodes[i].TimeStep)
		model.Logger1.Printf("%d,%d\n", i, nodes[i].TimeStep)
	}
}

func runNodeBLSSetup(node *bridge.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	node.MembershipKey = node.BlsSetup(wg)
}
