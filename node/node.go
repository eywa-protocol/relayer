package node

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/helpers"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/libp2p_pubsub"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/modelBLS"
	messageSigpb "github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/protobuf/messageWithSig"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
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
	"math/big"
	"strings"
	"sync"
	"time"
)

type Node struct {
	Ctx            context.Context
	Router         *mux.Router
	Server         bridges.Server
	DiscoveryPeers addrList
	EthClient_1    *ethclient.Client
	EthClient_2    *ethclient.Client
	pKey           *ecdsa.PrivateKey
	Host           host.Host
	Dht            *dht.IpfsDHT
	Service        *libp2p.Service
	P2PPubSub      *libp2p_pubsub.Libp2pPubSub
	NodeBLS        *modelBLS.Node
	Routing        routing.PeerRouting
	Discovery      *discovery.RoutingDiscovery
	BLSAddress     common.Address
	PrivKey        kyber.Scalar
}

type addrList []multiaddr.Multiaddr

//func (n Node) RunNode(wg *sync.WaitGroup) {
//	defer wg.Done()
//	n.NodeBLS.WaitForMsgNEW()
//}

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

func (n Node) runRPCService() (err error) {
	n.Service = libp2p.NewService(n.Host, protocol.ID("network-start"))
	err = n.Service.SetupRPC()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	n.Service.StartMessaging(n.Ctx, config.Config.TickerInterval)
	return
}

func (n Node) NewBridge() (srv *bridges.Server) {
	var bridgesList []bridges.Bridge
	ad, err := common2.NewDBridge(n.EthClient_1, "Health Chain 1", "1", common2.HealthFirst)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	ad2, err := common2.NewDBridge(n.EthClient_2, "Health Chain 2", "2", common2.HealthSecond)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	/*	ad3, err := common2.NewDBridge(n.EthClient_1, "SetMockPoolTestRequest", "test", common2.SetMockPoolTestRequestV2)
		if err != nil {
			logrus.Fatal(err)
			return
		}*/

	ad4, err := common2.NewDBridge(n.EthClient_1, "ChainlinkData", "control", common2.ChainlinkData)

	if err != nil {
		logrus.Fatal(err)
		return
	}

	bridgesList = append(bridgesList, ad)
	bridgesList = append(bridgesList, ad2)
	//bridgesList = append(bridgesList, ad3)
	bridgesList = append(bridgesList, ad4)
	srv = bridges.NewServer(bridgesList...)
	return
}

func (n Node) nodeExiats(blsAddr common.Address) bool {
	node, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	logrus.Trace(node.NodeWallet, node.BlsPubKey, node.NodeId, node.P2pAddress, node.BlsPointAddr)
	if err != nil || node.NodeWallet == common.HexToAddress("0") {
		return false
	}
	return true

}

func (n Node) GetPubKeysFromContract() (publicKeys []kyber.Point, err error) {
	suite := pairing.NewSuiteBn256()
	publicKeys = make([]kyber.Point, 0)
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
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

func (n Node) KeysFromFilesByConfigName(name string) (prvKey kyber.Scalar, blsAddr common.Address, err error) {

	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err = common2.ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return
	}
	blsAddr, err = common2.BLSAddrFromKeyFile(nodeKeyFile)
	if err != nil {
		return
	}
	return
}

func (n Node) NewBLSNode(topic *pubsub.Topic) (blsNode *modelBLS.Node, err error) {
	publicKeys, err := n.GetPubKeysFromContract()
	if err != nil {
		return
	}
	suite := pairing.NewSuiteBn256()
	if !n.nodeExiats(n.BLSAddress) {
		logrus.Errorf("node %x with keyFile %s does not exist", n.BLSAddress)

	} else {
		//logrus-Printf("BLS ADDRESS %v ", n.BLSAddress)
		node, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), n.BLSAddress)
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
				logrus.Tracef("len(topicParticipants) = [ %d ] len(n.DiscoveryPeers)/2+1 = [ %v ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.DiscoveryPeers)/2+1, len(n.Dht.RoutingTable().ListPeers()))
				if len(topicParticipants) >= len(n.DiscoveryPeers)/2+1 {
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
				//TODO disconnect from topic
				/** TODO:
				Is transaction true, otherwise repeate to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}

func (n *Node) ListenNodeOracleRequest(channel chan *wrappers.BridgeOracleRequest, wg *sync.WaitGroup) (err error) {
	//defer wg.Done()
	bridgeFilterer, err := wrappers.NewBridge(common.HexToAddress(config.Config.PROXY_NETWORK1), n.EthClient_1)
	if err != nil {
		return
	}
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		return
	}
	wg.Add(1)
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Fatalf("OracleRequest: %v", err)
			case event := <-channel:
				logrus.Trace("going to InitializePubSubWithTopicAndPeers")
				currentTopic := common2.ToHex(event.Raw.TxHash)
				logrus.Debugf("currentTopic %s", currentTopic)
				//n.P2PPubSub.RegisterTopicValidator("test", func(ctx context.Context, p peer.ID, msg *Message) bool {
				//	if string(msg.Data) == "invalid!" {
				//		return false
				//	} else {
				//		return true
				//	}
				//})
				sendTopic, err := n.P2PPubSub.JoinTopic(currentTopic)
				if err != nil {
					logrus.Fatal(err)
				}
				sub, err := sendTopic.Subscribe()
				logrus.Print(sub.Topic())
				logrus.Print(sendTopic.ListPeers())

				n.NodeBLS, err = n.NewBLSNode(sendTopic)
				if err != nil {
					logrus.Fatal(err)
				}
				if n.NodeBLS != nil {
					go n.StartProtocolByOracleRequest(event)

				}
			}
		}
	}()
	wg.Add(-1)
	return
}

func (n *Node) ListenNodeAddedEventInFirstNetwork() (err error) {
	bridgeFilterer, err := wrappers.NewNodeList(common.HexToAddress(config.Config.NODELIST_NETWORK1), n.EthClient_1)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.NodeListAddedNode)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchAddedNode(opt, channel)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Println("WatchAddedNode error:", err)
				break
			case event := <-channel:
				if event != nil {
					peerAddrFromEvent, err := multiaddr.NewMultiaddr(string(event.P2pAddress[:]))
					if err != nil {
						logrus.Error("WatchAddedNode error during prepare event into NewMultiaddr: ", err)
					}
					_ = peerAddrFromEvent
					logrus.Info("WatchAddedNode. Peer has been added: ", string(event.P2pAddress[:]))
					logrus.Printf("len(n.DiscoveryPeers) 1 =%d", len(n.DiscoveryPeers))
					n.DiscoveryPeers = append(n.DiscoveryPeers, peerAddrFromEvent)
					logrus.Printf("len(n.DiscoveryPeers) 2 =%d", len(n.DiscoveryPeers))
					n.Reconnect(n.DiscoveryPeers)
				}

			}
		}
	}()
	return
}

//func (n Node)getProtocolResult(resultCh chan string)  {
//	time.Sleep(2 * time.Second)
//	resultCh <-  n.StartProtocolByOracleRequest(n.CurrentRendezvous)
//}

func (n *Node) ReceiveRequestV2(event *wrappers.BridgeOracleRequest) (receipt *types.Receipt, err error) {
	privateKey, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		logrus.Error(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logrus.Error("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := n.EthClient_2.PendingNonceAt(n.Ctx, fromAddress)
	if err != nil {
		logrus.Error(err)
	}
	gasPrice, err := n.EthClient_2.SuggestGasPrice(n.Ctx)
	if err != nil {
		logrus.Error(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	instance, err := wrappers.NewBridge(common.HexToAddress(config.Config.PROXY_NETWORK2), n.EthClient_2)
	if err != nil {
		logrus.Error(err)
	}

	/**   TODO: apporove that tx was real  */
	/**   TODO: apporove that tx was real  */
	/**   TODO: apporove that tx was real  */
	/**   TODO: apporove that tx was real  */

	oracleRequest := helpers.OracleRequest{
		RequestType:    event.RequestType,
		Bridge:         event.Bridge,
		RequestId:      event.RequestId,
		Selector:       event.Selector,
		ReceiveSide:    event.ReceiveSide,
		OppositeBridge: event.OppositeBridge,
	}
	/** Invoke bridge on another side */
	tx, err := instance.ReceiveRequestV2(auth, "", nil, oracleRequest.Selector, [32]byte{}, oracleRequest.ReceiveSide)
	if err != nil {
		logrus.Error(err)
	}
	receipt, err = helpers.WaitTransactionWithRetry(n.EthClient_2, tx)
	if err != nil || receipt == nil {
		return nil, errors.New(fmt.Sprintf("ReceiveRequestV2 Failed %v", err.Error()))
	}

	return

}

func (n Node) Reconnect(bootstrapPeers []multiaddr.Multiaddr) {
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		if n.Host.ID().Pretty() != peerinfo.ID.Pretty() && n.Host.Network().Connectedness(peerinfo.ID) != network.Connected {
			ctx, cancel := context.WithTimeout(n.Ctx, time.Second*10)
			defer cancel()
			for {
				err := n.Host.Connect(ctx, *peerinfo)
				if err != nil {
					logrus.Errorf("Error while connecting to node %q: %-v", peerinfo, err)
				} else {
					logrus.Infof("Connection established with node: %q", peerinfo)
					break
				}
				if ctx.Err() != nil {
					logrus.Error(ctx.Err())
					break
				}
				time.Sleep(2 * time.Second)
				fmt.Println("Trying to connect after unsuccessful connect...")
			}
		}
	}
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
	logrus.Printf("Announcing ourselves...")
	var routingDiscovery = discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, rendezvous)
	logrus.Printf("Successfully announced!")

	ticker := time.NewTicker(config.Config.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logrus.Tracef("Looking advertised peers by rendezvous....")
			//TODO: enhance, because synchronous
			peers, err := discovery.FindPeers(n.Ctx, routingDiscovery, rendezvous)
			if err != nil {
				logrus.Fatal(err)
			}
			ctn := 0
			for _, p := range peers {
				if p.ID == n.Host.ID() {
					continue
				}
				logrus.Tracef("Discovery: FoundedPee %v, isConnected: %s", p, n.Host.Network().Connectedness(p.ID) == network.Connected)
				//TODO: add into if: "&&  n.n.DiscoveryPeers.contains(p.ID)"
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
