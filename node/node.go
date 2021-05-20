package node

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

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
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/linkpoolio/bridges"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/util/encoding"
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
	discovery.Advertise(n.Ctx, n.Discovery, n.NodeBLS.CurrentRendezvous)
	go n.NodeBLS.AdvanceWithTopic(0, n.NodeBLS.CurrentRendezvous)
	wg.Add(1)
	go n.NodeBLS.WaitForMsgNEW(consensuChannel)
	consensus := <-consensuChannel
	executed := false
	if consensus == true && executed == false {

		logrus.Tracef("LEADER id %s my ID %s", n.NodeBLS.Leader.Pretty(), n.Host.ID().Pretty())
		if n.NodeBLS.Leader.Pretty() == n.Host.ID().Pretty() {
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
	wg.Wait()
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

func (n Node) initNewPubSub(topic string) (p2pPubSub *libp2p_pubsub.Libp2pPubSub) {
	p2pPubSub = new(libp2p_pubsub.Libp2pPubSub)
	p2pPubSub.InitializePubSubWithTopic(n.Host, topic)
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

func (n Node) NewBLSNode(topic string) (blsNode *modelBLS.Node, err error) {
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
		n.Dht.RefreshRoutingTable()
		blsNode = func() *modelBLS.Node {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			mychannel := make(chan bool)
			for {
				select {
				case <-ticker.C:
					n.initNewPubSub(topic)
					go n.DiscoverWithEvtTopic(topic)
					topicParticipants := n.P2PPubSub.ListPeersByTopic(topic)
					topicParticipants = append(topicParticipants, n.Host.ID())
					if len(topicParticipants) > 6 {
						logrus.Tracef("Starting Leader election !!!")
						leaderPeerId, err := libp2p.RelayerLeaderNode(topic, topicParticipants)
						if err != nil {
							panic(err)
						}
						logrus.Warnf("LEADER IS %v", leaderPeerId)
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
							CurrentRendezvous: topic,
							Leader:            leaderPeerId,
						}
						ticker.Stop()
						return blsNode
					} else {
						//TODO:
						//this place leaks after at least one e2e test. If test was executed twice the leaks double
						// top -o RES # the pid of most active node
						// docker ps -q | xargs docker inspect --format '{{.State.Pid}}, {{.Name}}'
						// checks out the names of docker nodes through commands above.
						logrus.Info("topicParticipants: ", topicParticipants)
					}

				case <-mychannel:
					return nil
				}
			}
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
				logrus.Tracef("ReceiveRequest: %v %v %v", event.ReqId, event.ReceiveSide, common2.ToHex(event.Tx))
				if n.P2PPubSub != nil {
					n.P2PPubSub.Disconnect()
				}

				/** TODO:
				Is transaction true, otherwise repeate to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}

func (n *Node) ListenNodeOracleRequest() (oracleRequest *helpers.OracleRequest, err error) {

	bridgeFilterer, err := wrappers.NewBridge(common.HexToAddress(config.Config.PROXY_NETWORK1), n.EthClient_1)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeOracleRequest)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Fatalf("OracleRequest: %v", err)
			case event := <-channel:
				currentRendezvous := common2.ToHex(event.Raw.TxHash)
				n.P2PPubSub = n.initNewPubSub(currentRendezvous)

				n.NodeBLS, err = n.NewBLSNode(currentRendezvous)
				if err != nil {
					logrus.Fatal(err)
				}
				if n.NodeBLS != nil {
					go n.DiscoverWithEvtTopic(n.NodeBLS.CurrentRendezvous)
					go n.StartProtocolByOracleRequest(event)

				}
			}
		}
	}()
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
			case event := <-channel:
				if event != nil {
					peerAddrFromEvent, err := multiaddr.NewMultiaddr(string(event.P2pAddress[:]))
					if err != nil {
						n.DiscoveryPeers = append(n.DiscoveryPeers, peerAddrFromEvent)
					}
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

func (n Node) DiscoverWithEvtTopic(topic string) {
	var routingDiscovery = discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, topic)
	ticker := time.NewTicker(config.Config.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//n.FindPeers()
			peers, err := discovery.FindPeers(n.Ctx, routingDiscovery, topic)
			if err != nil {
				logrus.Fatal(err)
			}

			peerIds := n.P2PPubSub.ListPeersByTopic(topic)
			for _, pID := range peerIds {
				if pID == n.Host.ID() {
					continue
				}

				if n.Host.Network().Connectedness(pID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, pID)
					logrus.Tracef("Connected to peer %s", pID.Pretty())

					if err != nil {
						logrus.Errorf("Host.Network().Connectedness PROBLEM: %v", err)
						continue
					}
				}
			}

			for _, p := range peers {
				if p.ID == n.Host.ID() {
					continue
				}
				if n.Host.Network().Connectedness(p.ID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, p.ID)
					logrus.Infof("Connected to peer %s", p.ID.Pretty())
					if err != nil {
						logrus.Errorf("Host.Network().Connectedness PROBLEM: %v", err)
						continue
					}
				}
			}
		case <-n.Ctx.Done():
			return
		}
	}
}
