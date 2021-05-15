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
	"math/big"
	"strings"
	"sync"
	"time"
)

type Node struct {
	Ctx               context.Context
	Router            *mux.Router
	Server            bridges.Server
	DiscoveryPeers    addrList
	CurrentRendezvous string
	EthClient_1       *ethclient.Client
	EthClient_2       *ethclient.Client
	pKey              *ecdsa.PrivateKey
	Host              host.Host
	Dht               *dht.IpfsDHT
	Service           *libp2p.Service
	P2PPubSub         *libp2p_pubsub.Libp2pPubSub
	NodeBLS           *modelBLS.Node
	Routing           routing.PeerRouting
	Discovery         *discovery.RoutingDiscovery
	PublicKeys        []kyber.Point
	BLSAddress        common.Address
	PrivKey           kyber.Scalar
}

type addrList []multiaddr.Multiaddr

//func (n Node) RunNode(wg *sync.WaitGroup) {
//	defer wg.Done()
//	n.NodeBLS.WaitForMsgNEW()
//}

func (n Node) StartProtocolByOracleRequest(event *wrappers.BridgeOracleRequest) {
	n.NodeBLS, _ = n.NewBLSNode()

	consensuChannel := make(chan bool)
	wg := &sync.WaitGroup{}
	defer wg.Done()
	wg.Add(1)
	discovery.Advertise(n.Ctx, n.Discovery, n.CurrentRendezvous)
	go n.NodeBLS.AdvanceWithTopic(0, n.CurrentRendezvous)
	wg.Add(1)
	go n.NodeBLS.WaitForMsgNEW(consensuChannel)
	consensus := <-consensuChannel
	executed := false
	if consensus == true && executed == false {
		logrus.Println("Call external contract method test")
		recept, err := n.ReceiveRequestV2(event)
		if err != nil {
			logrus.Errorf("%v", err)
		}
		if recept != nil {
			executed = true
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

func (n Node) initNewPubSub() (p2pPubSub *libp2p_pubsub.Libp2pPubSub) {
	p2pPubSub = new(libp2p_pubsub.Libp2pPubSub)
	return
}

func (n Node) nodeExiats(blsAddr common.Address) bool {
	node, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	logrus.Print(node.NodeWallet, node.BlsPubKey, node.NodeId, node.P2pAddress, node.BlsPointAddr)
	if err != nil || node.NodeWallet == common.HexToAddress("0") {
		return false
	}
	return true

}

func (n Node) SetPublicKeys() (publicKeys []kyber.Point, err error) {
	publicKeys = make([]kyber.Point, 0)
	suite := pairing.NewSuiteBn256()
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		blsPubKey := string(node.BlsPubKey[:])
		p, err := encoding.ReadHexPoint(suite, strings.NewReader(blsPubKey))
		if err != nil {
			panic(err)
		}
		publicKeys = append(publicKeys, p)
	}
	return
}

func (n Node) KeysFromFilesByConfigName(name string) (prvKey kyber.Scalar, blsAddr common.Address, err error) {

	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err = common2.ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return
	}
	blsAddr = common2.BLSAddrFromKeyFile(nodeKeyFile)
	return
}

func (n Node) NewBLSNode() (blsNode *modelBLS.Node, err error) {

	suite := pairing.NewSuiteBn256()
	if !n.nodeExiats(n.BLSAddress) {
		logrus.Errorf("node %x with keyFile %s does not exist", n.BLSAddress)

	} else {
		//logrus.Printf("BLS ADDRESS %v ", n.BLSAddress)
		node, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), n.BLSAddress)
		if err != nil {
			return nil, err
		}

		mask, err := sign.NewMask(suite, n.PublicKeys, nil)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("HostId %v\n%v\n%v\n%v\n%v\n", node.BlsPointAddr, string(node.P2pAddress), node.NodeId, string(node.BlsPubKey), node.NodeWallet)
		fmt.Printf("---------------->NODE ID %d %v\n", node.NodeId, int(node.NodeId))

		//n.Dht.RefreshRoutingTable()
		//n.Dht.RoutingTable().Print()
		peerForConsensus := n.P2PPubSub.ListPeersByTopic(n.CurrentRendezvous)

		if len(peerForConsensus) == 0 {
			logrus.Error(errors.New("no peers in table"))
		}
		consensusCount := len(peerForConsensus)
		if consensusCount < 3 {
			logrus.Warnf("consensusCount: %d IS LESS THEN THREE NODES !!!", consensusCount)
			consensusCount = 3
		}
		logrus.Printf("\\\\\\\\\\\\\\\\\\\\\\ consensusCount %d //////////////////////////////////", consensusCount)

		blsNode = &modelBLS.Node{
			Id:           int(node.NodeId),
			TimeStep:     0,
			ThresholdWit: consensusCount/2 + 1,
			ThresholdAck: consensusCount/2 + 1,
			Acks:         0,
			ConvertMsg:   &messageSigpb.Convert{},
			Comm:         n.P2PPubSub,
			History:      make([]modelBLS.MessageWithSig, 0),
			Signatures:   make([][]byte, len(n.PublicKeys)),
			SigMask:      mask,
			PublicKeys:   n.PublicKeys,
			PrivateKey:   n.PrivKey,
			Suite:        suite,
		}
	}
	return
}

func (n *Node) ListenNodeOracleRequest() (err error) {

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
				logrus.Println("OracleRequest error:", err)
			case event := <-channel:
				//logrus.Printf("OracleRequest %v %v id: %v\n", event.RequestType, common2.ToHex(event.RequestId), common2.ToHex(event.RequestId))
				n.CurrentRendezvous = common2.ToHex(event.Raw.TxHash)
				n.P2PPubSub = n.initNewPubSub()
				n.P2PPubSub.InitializePubSubWithTopic(n.Host,n.CurrentRendezvous )
				go n.Discover()
				if event != nil {
					go n.StartProtocolByOracleRequest(event)
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

	instance, err := wrappers.NewBridge(common.HexToAddress(config.Config.PROXY_NETWORK1), n.EthClient_2)
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
	receipt, err = helpers.WaitTransaction(n.EthClient_2, tx)
	if err != nil || receipt == nil {
		return nil, errors.New(fmt.Sprintf("ReceiveRequestV2 Failed %v", err.Error()))
	}

	return

}

func (n Node) Discover() {

	logrus.Print("CurrentRendezvous ", n.CurrentRendezvous)
	n.P2PPubSub.InitializePubSubWithTopic(n.Host, n.CurrentRendezvous)
	var routingDiscovery = discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, n.CurrentRendezvous)
	ticker := time.NewTicker(config.Config.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.Ctx.Done():
			return
		case <-ticker.C:
			n.FindPeers()
			peers, err := discovery.FindPeers(n.Ctx, routingDiscovery, n.CurrentRendezvous)
			if err != nil {
				logrus.Fatal(err)
			}

			peerIds := n.P2PPubSub.ListPeersByTopic(n.CurrentRendezvous)
			for _, pID := range peerIds {
				//logrus.Printf("discovery.FindPeers PEER %v  - - - - - - - - - - - - - - >", p.Addrs)
				if pID == n.Host.ID() {
					continue
				}
				if n.Host.Network().Connectedness(pID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, pID)
					logrus.Printf("!!!!!!!!!!!!!!!!!!!!!! Connected to peer %s ! ! ! ! ! ! !\n", pID.Pretty())

					if err != nil {
						logrus.Errorf("Host.Network().Connectedness PROBLEM: %v", err)
						continue
					}
				}
			}

			for _, p := range peers {
				//logrus.Printf("discovery.FindPeers PEER %v  - - - - - - - - - - - - - - >", p.Addrs)
				if p.ID == n.Host.ID() {
					continue
				}
				if n.Host.Network().Connectedness(p.ID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, p.ID)
					logrus.Printf("!!!!!!!!!!!!!!!!!!!!!! Connected to peer %s !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!111\n", p.ID.Pretty())
					if err != nil {
						logrus.Errorf("Host.Network().Connectedness PROBLEM: %v", err)
						continue
					}
				}
			}
		}
	}
}

func (n Node) FindPeers() {
	//logrus.Printf("Начинаю поиcк узлов\n")
	peerInfos, err := discovery.FindPeers(n.Ctx, n.Discovery, n.CurrentRendezvous)
	if len(peerInfos) != 0 {
		discovery.Advertise(n.Ctx, n.Discovery, n.CurrentRendezvous)
	} else {
		return
	}
	if len(peerInfos) > 1 {
		logrus.Printf("нвйдено узлов: %v", len(peerInfos))
	}

	for i, pe := range peerInfos {
		logrus.Printf("(%d):%s\n", i, pe.ID.Pretty())
	}
	if err != nil {
		panic(err)
	}
	for _, peerInfo := range peerInfos {
		//logrus.Printf("%v нашел узел: %s\n",n.Host.ID().Pretty(),  peerInfo.ID.Pretty())
		if peerInfo.ID == n.Host.ID() {
			continue
		}
		logrus.Printf("Пробую подключиться:%s\n", peerInfo.ID.Pretty())
		if err := n.Host.Connect(n.Ctx, peerInfo); err != nil {
			logrus.Printf("Узел: %s\n", peerInfo.ID.Pretty())
			continue
		}
		logrus.Printf("Успешно подключен:%s\n", peerInfo.ID.Pretty())
	}
	time.Sleep(time.Minute * 2)
}
