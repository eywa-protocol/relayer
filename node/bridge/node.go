package bridge

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/helpers"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	"github.com/digiu-ai/p2p-bridge/libp2p/pub_sub_bls/libp2p_pubsub"
	"github.com/digiu-ai/p2p-bridge/libp2p/pub_sub_bls/modelBLS"
	messageSigpb "github.com/digiu-ai/p2p-bridge/libp2p/pub_sub_bls/protobuf/messageWithSig"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/linkpoolio/bridges"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/util/encoding"
)

type Node struct {
	Ctx       context.Context
	Router    *mux.Router
	Server    bridges.Server
	Host      host.Host
	Dht       *dht.IpfsDHT
	Service   *libp2p.Service
	P2PPubSub *libp2p_pubsub.Libp2pPubSub
	NodeBLS   *modelBLS.Node
	Routing   routing.PeerRouting
	Discovery *discovery.RoutingDiscovery
	PrivKey   kyber.Scalar
	Client1   Client
	Client2   Client
	Client3   Client
}

type Client struct {
	ethClient        *ethclient.Client
	ECDSA_KEY        string
	Bridge           wrappers.BridgeSession
	BridgeFilterer   wrappers.BridgeFilterer
	NodeList         wrappers.NodeListSession
	NodeListFilterer wrappers.NodeListFilterer
}

type addrList map[peer.ID]multiaddr.Multiaddr

// func (n Node) RunNode(wg *sync.WaitGroup) {
//	defer wg.Done()
//	n.NodeBLS.WaitForMsgNEW()
// }

func (n Node) StartProtocolByOracleRequest(event *wrappers.BridgeOracleRequest) {
	consensuChannel := make(chan bool)
	wg := &sync.WaitGroup{}
	defer wg.Done()
	wg.Add(1)

	logrus.Tracef("сurrentRendezvous %v LEADER %v", n.NodeBLS.CurrentRendezvous, n.NodeBLS.Leader)

	go n.NodeBLS.AdvanceWithTopic(0, n.NodeBLS.CurrentRendezvous)
	wg.Add(1)
	go n.NodeBLS.WaitForMsgNEW(consensuChannel)
	consensus := <-consensuChannel
	executed := false
	if consensus == true && executed == false {
		logrus.Tracef("Starting Leader election !!!")
		leaderPeerId, err := libp2p.RelayerLeaderNode(n.NodeBLS.CurrentRendezvous, n.NodeBLS.Participants)
		if err != nil {
			panic(err)
		}
		logrus.Infof("LEADER IS %v", leaderPeerId)
		logrus.Debugf("LEADER id %s my ID %s", n.NodeBLS.Leader.Pretty(), n.Host.ID().Pretty())
		if leaderPeerId.Pretty() == n.Host.ID().Pretty() {
			logrus.Info("LEADER going to Call external chain contract method")
			recept, err := n.ReceiveRequestV2(event)
			if err != nil {
				logrus.Errorf("%v", err)
			}
			if recept != nil {
				executed = true
			}
		}
	}
	logrus.Println("The END of Protocol")
}

func (n Node) nodeExists(nodeIdAddress common.Address) bool {
	node, err := common2.GetNode(n.Client1.ethClient, common.HexToAddress(config.Config.NODELIST_NETWORK1), nodeIdAddress)
	if err != nil || node.NodeWallet == common.HexToAddress("0") {
		return false
	}
	return true

}

func (n Node) GetPubKeysFromContract() (publicKeys []kyber.Point, err error) {
	suite := pairing.NewSuiteBn256()
	publicKeys = make([]kyber.Point, 0)
	nodes, err := common2.GetNodesFromContract(n.Client1.ethClient, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		p, err := encoding.ReadHexPoint(suite, strings.NewReader(string(node.BlsPubKey)))
		if err != nil {
			panic(err)
		}
		publicKeys = append(publicKeys, p)
	}
	return
}

func (n Node) AddPubkeyToNodeKeys(blsPubKey []byte) {
	suite := pairing.NewSuiteBn256()
	blsPKey := string(blsPubKey[:])
	p, err := encoding.ReadHexPoint(suite, strings.NewReader(blsPKey))
	if err != nil {
		panic(err)
	}
	n.NodeBLS.PublicKeys = append(n.NodeBLS.PublicKeys, p)
}

func (n Node) KeysFromFilesByConfigName(name string) (prvKey kyber.Scalar, err error) {

	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err = common2.ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return
	}

	return
}

func (n Node) NewBLSNode(topic *pubsub.Topic, client Client) (blsNode *modelBLS.Node, err error) {
	publicKeys, err := n.GetPubKeysFromContract()
	if err != nil {
		return
	}

	suite := pairing.NewSuiteBn256()

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	if !n.nodeExists(nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())

	} else {
		logrus.Tracef("Host.ID() %v ", n.Host.ID())
		node, err := client.NodeList.GetNode(nodeIdAddress)
		if err != nil {
			return nil, err
		}

		mask, err := sign.NewMask(suite, publicKeys, nil)
		if err != nil {
			return nil, err
		}
		blsNode = func() *modelBLS.Node {
			ctx, cancel := context.WithDeadline(n.Ctx, time.Now().Add(3*time.Second))
			defer cancel()
			for {

				topicParticipants := topic.ListPeers()
				topicParticipants = append(topicParticipants, n.Host.ID())
				// logrus.Tracef("len(topicParticipants) = [ %d ] len(n.DiscoveryPeers)/2+1 = [ %v ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants) /*, len(n.DiscoveryPeers)/2+1*/, len(n.Dht.RoutingTable().ListPeers()))
				logrus.Tracef("len(topicParticipants) = [ %d ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.Dht.RoutingTable().ListPeers()))
				if len(topicParticipants) >= 5 {
					blsNode = &modelBLS.Node{
						Id:                int(node.NodeId),
						TimeStep:          0,
						ThresholdWit:      len(topicParticipants)/2 + 1,
						ThresholdAck:      len(topicParticipants)/2 + 1,
						Acks:              0,
						ConvertMsg:        &messageSigpb.Convert{},
						Comm:              n.P2PPubSub,
						History:           make([]modelBLS.MessageWithSig, 0),
						Signatures:        make([][]byte, len(publicKeys)),
						SigMask:           mask,
						PublicKeys:        publicKeys,
						PrivateKey:        n.PrivKey,
						Suite:             suite,
						Participants:      topicParticipants,
						CurrentRendezvous: topic.String(),
						Leader:            "",
					}
					break
				}
				if ctx.Err() != nil {
					logrus.Warnf("Not enaugh participants %d , %v", len(topicParticipants), ctx.Err())
					_ = topic.Close()
					break
				}
				time.Sleep(300 * time.Millisecond)
			}

			return blsNode
		}()

	}
	return
}

func (n *Node) ListenReceiveRequest(clientNetwork *ethclient.Client, proxyNetwork common.Address) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork, clientNetwork)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeReceiveRequest)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchReceiveRequest(opt, channel)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case _ = <-sub.Err():
				break
			case event := <-channel:
				logrus.Infof("ReceiveRequest: %v %v %v", event.ReqId, event.ReceiveSide, common2.ToHex(event.Tx))
				// TODO disconnect from topic
				/** TODO:
				Is transaction true, otherwise repeate to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}

func (n *Node) ListenNodeOracleRequest(channel chan *wrappers.BridgeOracleRequest, wg *sync.WaitGroup, client Client) (err error) {
	opt := &bind.WatchOpts{}
	sub, err := client.BridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		return
	}
	wg.Add(1)
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Errorf("OracleRequest: %v", err)
				return
			case event := <-channel:
				logrus.Info("going to InitializePubSubWithTopicAndPeers")
				currentTopic := common2.ToHex(event.Raw.TxHash)
				logrus.Debugf("currentTopic %s", currentTopic)
				sendTopic, err := n.P2PPubSub.JoinTopic(currentTopic)
				if err != nil {
					logrus.Error("JoinTopic ", err)
				}
				defer sendTopic.Close()
				if sendTopic != nil {
					sub, err := sendTopic.Subscribe()
					if err != nil {
						logrus.Error("Subscribe ", err)
					}
					defer sub.Cancel()
					logrus.Print(sub.Topic())
					logrus.Print(sendTopic.ListPeers())

					n.NodeBLS, err = n.NewBLSNode(sendTopic, client)
					if err != nil {
						logrus.Error("NewBLSNode ", err)
					}
					if n.NodeBLS != nil {
						go n.StartProtocolByOracleRequest(event)

					}
				}
			}
		}
	}()
	// wg.Add(-1)
	return
}

func (n *Node) ReceiveRequestV2(event *wrappers.BridgeOracleRequest) (receipt *types.Receipt, err error) {

	client, err := n.GetNodeClientByChainId(event.Chainid)

	/** Invoke bridge on another side */

	tx, err := client.Bridge.ReceiveRequestV2(event.RequestId, event.Selector, event.ReceiveSide, event.Bridge)

	if err != nil {
		logrus.Error(err)
	}
	if tx != nil {
		receipt, err = helpers.WaitTransaction(client.ethClient, tx)
		if err != nil || receipt == nil {
			return nil, errors.New(fmt.Sprintf("ReceiveRequestV2 Failed %v", err))
		}
	}
	return
}

func (n Node) Discover(topic string) {
	go libp2p.Discover(n.Ctx, n.Host, n.Dht, topic, 3*time.Second)
}

func (n Node) InitializeCoomonPubSub() (p2pPubSub *libp2p_pubsub.Libp2pPubSub) {
	return new(libp2p_pubsub.Libp2pPubSub)
}

/**
* 	Announce your presence in network using a rendezvous point
*	With the DHT set up, it’s time to discover other peers
*
*	The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
*	It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.
*	routingDiscovery.Advertise makes this node announce that it can provide a value for the given key.
*	Where a key in this case is rendezvousString. Other peers will hit the same key to find other peers.
 */
func (n Node) DiscoverByRendezvous(rendezvous string) {

	//	The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
	//	It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.

	// TODO: When TTL elapsed should check precense in network
	// QEST: What's happening with presence in network when node goes down (in DHT table, in while other nodes is trying to connect)
	logrus.Printf("Announcing ourselves with rendezvous [%s] ...", rendezvous)
	var routingDiscovery = discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, rendezvous)
	logrus.Printf("Successfully announced! n.Host.ID():%s ", n.Host.ID().Pretty())

	ticker := time.NewTicker(config.Config.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logrus.Infof("Looking advertised peers by rendezvous %s....", rendezvous)
			// TODO: enhance, because synchronous
			peers, err := discovery.FindPeers(n.Ctx, routingDiscovery, rendezvous)
			if err != nil {
				logrus.Fatal("FindPeers ", err)
			}
			ctn := 0
			for _, p := range peers {
				if p.ID == n.Host.ID() {
					continue
				}
				logrus.Tracef("Discovery: FoundedPee %v, isConnected: %t", p, n.Host.Network().Connectedness(p.ID) == network.Connected)
				// TODO: add into if: "&&  n.n.DiscoveryPeers.contains(p.ID)"

				if n.Host.Network().Connectedness(p.ID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, p.ID)
					if err != nil {
						logrus.Tracef("Connect to peer was unsuccessful: %s", p.ID)

					} else {
						logrus.Tracef("Connected to peer %s", p.ID.Pretty())
					}
				}
				if n.Host.Network().Connectedness(p.ID) == network.Connected {
					ctn++
				}

			}
			logrus.Infof("Total amout discovered and connected peers: %d", ctn)

		case <-n.Ctx.Done():
			return
		}
	}
}