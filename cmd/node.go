package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/libp2p_pubsub"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/modelBLS"
	messageSigpb "github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/protobuf/messageWithSig"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/linkpoolio/bridges"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/util/encoding"
	"os"
	"os/signal"
	p "path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
}

type addrList []multiaddr.Multiaddr

func main() {
	var mode string
	var path string
	flag.StringVar(&mode, "mode", "serve", "relayer mode. Default is serve")
	flag.StringVar(&path, "cnf", "config.env", "config file absolute path")
	flag.Parse()
	logrus.Printf("mode %v path %v", mode, path)
	file := filepath.Base(path)
	fname := strings.TrimSuffix(file, p.Ext(file))
	logrus.Println("FILE", fname)
	if mode == "init" {
		err := nodeInit(path, fname)
		if err != nil {
			logrus.Fatalf("nodeInit %v", err)
			panic(err)
		}

	} else {
		err := NewNode(path, fname)
		if err != nil {
			logrus.Fatalf("NewNode %v", err)
			panic(err)
		}

	}

}

func NewNode(path, name string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx:               ctx,
		CurrentRendezvous: "FirstRun",
	}

	n.pKey, err = common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Printf("n.Config.PORT_1 %d", config.Config.PORT_1)

	n.EthClient_1, n.EthClient_2, err = getEthClients()

	if err != nil {
		return
	}
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		logrus.Printf(string(node.P2pAddress))
	}

	var bootstrapPeers []multiaddr.Multiaddr
	suite := pairing.NewSuiteBn256()

	nodesPubKeys := make([]kyber.Point, 0)

	for i, node := range nodes {
		logrus.Printf("Node:%d\n %v\n %v\n %v\n %v\n", i, node.Enable, node.NodeWallet, string(node.P2pAddress[:]), string(node.BlsPubKey[:]))
		peerMA, err := multiaddr.NewMultiaddr(string(node.P2pAddress[:]))
		if err != nil {
			return err
		}
		bootstrapPeers = append(bootstrapPeers, peerMA)
		blsPubKey := string(node.BlsPubKey[:])
		logrus.Printf("BlsPubKey %v", blsPubKey)
		p, err := encoding.ReadHexPoint(suite, strings.NewReader(blsPubKey))
		if err != nil {
			panic(err)
		}
		nodesPubKeys = append(nodesPubKeys, p)
	}

	for _, peer := range bootstrapPeers {
		logrus.Printf("peer multyAddress %v", peer)
	}

	key_file := "keys/" + name + "-ecdsa.key"

	n.Host, err = libp2p.NewHost(n.Ctx, 0, key_file, config.Config.P2P_PORT)
	if err != nil {
		return
	}
	//_ = libp2p.WriteHostAddrToConfig(n.Host, "keys/peer.env")
	n.Dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers)

	n.NodeBLS, err = n.newBLSNode(name, nodesPubKeys)
	if err != nil {
		logrus.Errorf("newBLSNode %v", err)
		return err
	}

	//n.P2PPubSub.StartBLSNode(n.NodeBLS, 1, 1)
	StartTest(n.NodeBLS, 2, 1)
	go libp2p.Discover(n.Ctx, n.Host, n.Dht, "TLC", config.Config.TickerInterval)
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	n.Server.Start(config.Config.PORT_1)
	run(n.Host, cancel)
	return
}

//func StartTest(bls *modelBLS.Node, i int, i2 int) {
//
//}

func runNode(node *modelBLS.Node, stop int, wg *sync.WaitGroup) {
	defer wg.Done()
	node.WaitForMsg(stop)
}

// StartTest is used for starting tlc nodes
func StartTest(node *modelBLS.Node, stop int, fails int) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go node.Advance(0)
	wg.Add(1)
	go runNode(node, stop, wg)
	wg.Wait()
	fmt.Println("The END")
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

func getEthClients() (c1 *ethclient.Client, c2 *ethclient.Client, err error) {
	logrus.Printf("config.Config.NETWORK_RPC_1 %s", config.Config.NETWORK_RPC_1)
	c1, err = ethclient.Dial(config.Config.NETWORK_RPC_1)
	if err != nil {
		return
	}

	c2, err = ethclient.Dial(config.Config.NETWORK_RPC_2)
	if err != nil {
		return
	}
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

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logrus.Printf("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func (n Node) newBLSNode(name string, publicKeys []kyber.Point) (*modelBLS.Node, error) {
	n.P2PPubSub = new(libp2p_pubsub.Libp2pPubSub)
	n.P2PPubSub.InitializePubSub(n.Host)
	suite := pairing.NewSuiteBn256()
	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err := common2.ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return nil, err
	}

	blsAddr := common2.BLSAddrFromKeyFile(nodeKeyFile)

	logrus.Printf("BLS ADDRESS %v ", blsAddr)
	node, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	if err != nil {
		return nil, err
	}

	mask, err := sign.NewMask(suite, publicKeys, nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("HostId %v\n%v\n%v\n%v\n%v\n", node.BlsPointAddr, node.P2pAddress, node.NodeId, node.BlsPubKey, node.NodeWallet)
	fmt.Printf("---------------->NODE ID %d %v\n", node.NodeId, int(node.NodeId))

	return &modelBLS.Node{
		Id:           int(node.NodeId),
		TimeStep:     0,
		ThresholdWit: len(publicKeys)/2 + 1,
		ThresholdAck: len(publicKeys)/2 + 1,
		Acks:         0,
		ConvertMsg:   &messageSigpb.Convert{},
		Comm:         n.P2PPubSub,
		History:      make([]modelBLS.MessageWithSig, 0),
		Signatures:   make([][]byte, len(publicKeys)),
		SigMask:      mask,
		PublicKeys:   publicKeys,
		PrivateKey:   prvKey,
		Suite:        suite,
	}, nil

}

func loadNodeConfig(path string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Printf("started in directory %s", dir)
	err = config.LoadConfigAndArgs(path)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	return
}

func nodeInit(path, name string) (err error) {
	logrus.Print("nodeInit START")

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	blsAddr, pub, err := common2.CreateBN256Key(name)
	if err != nil {
		return
	}

	logrus.Printf("pubkey %v", pub)

	err = common2.GenECDSAKey(name)
	if err != nil {
		return
	}

	logrus.Printf("keyfile %v port %v", "keys/"+name+"-ecdsa.key", config.Config.P2P_PORT)
	h, err := libp2p.NewHost(context.Background(), 0, "keys/"+name+"-ecdsa.key", config.Config.P2P_PORT)
	if err != nil {
		return
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, "keys/"+name+"-peer.env")
	c1, c2, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Printf("nodelist 1 %v blsAddress %v", common.HexToAddress(config.Config.ECDSA_KEY_1), pub)
	pKey1, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c1, pKey1, common.HexToAddress(config.Config.NODELIST_NETWORK1), common.HexToAddress(config.Config.ECDSA_KEY_1), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network1 %v", err)
	}
	common2.PrintNodes(c1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	logrus.Printf("nodelist 2 %v", common.HexToAddress(config.Config.NODELIST_NETWORK2))

	pKey2, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c2, pKey2, common.HexToAddress(config.Config.NODELIST_NETWORK2), common.HexToAddress(config.Config.ECDSA_KEY_2), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network2 %v", err)
	}

	common2.PrintNodes(c2, common.HexToAddress(config.Config.NODELIST_NETWORK2))
	return
}
