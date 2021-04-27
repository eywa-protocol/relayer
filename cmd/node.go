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
	"os"
	"os/signal"
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
	BRIDGE_1_ADDRESS  common.Address
	ORACLE_1_ADDRESS  common.Address
	BRIDGE_2_ADDRESS  common.Address
	ORACLE_2_ADDRESS  common.Address
	pKey              *ecdsa.PrivateKey
	Host              host.Host
	Dht               *dht.IpfsDHT
	Service           *libp2p.Service
	P2PPubSub         *libp2p_pubsub.Libp2pPubSub
	NodeBLS           modelBLS.Node
}

type addrList []multiaddr.Multiaddr

func NewNode(path string) (err error) {

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

	err = n.initEthClients()
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Printf("n.Config.PORT_1 %d", config.Config.PORT_1)
	var bootstrapPeers []multiaddr.Multiaddr
	if len(config.Config.BOOTSTRAP_PEER) > 0 {
		ma, err := multiaddr.NewMultiaddr(config.Config.BOOTSTRAP_PEER)
		if err != nil {
			return err
		}
		bootstrapPeers = append(bootstrapPeers, ma)
	}

	n.Host, err = libp2p.NewHost(n.Ctx, 0, config.Config.KEY_FILE, config.Config.P2P_PORT)
	if err != nil {
		return
	}
	_ = libp2p.WriteHostAddrToConfig(n.Host, "keys/peer.env")
	n.Dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers)

	n.NodeBLS = *n.newBLSNode(4)

	n.P2PPubSub.StartBLSNode(&n.NodeBLS, 1, 0)

	go libp2p.Discover(n.Ctx, n.Host, n.Dht, "test", config.Config.TickerInterval)
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	n.Server.Start(config.Config.PORT_1)
	run(n.Host, cancel)
	return
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

func (n Node) initEthClients() (err error) {
	logrus.Printf("config.Config.NETWORK_RPC_1 %s", config.Config.NETWORK_RPC_1)
	n.EthClient_1, err = ethclient.Dial(config.Config.NETWORK_RPC_1)
	if err != nil {
		return
	}

	n.EthClient_2, err = ethclient.Dial(config.Config.NETWORK_RPC_2)
	if err != nil {
		return
	}
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

func main() {
	var mode string
	var path string
	flag.StringVar(&mode, "mode", "serve", "Specify mode. Default is serve")
	flag.StringVar(&path, "cnf", ".", "config file absolute path")
	flag.Usage = func() {
		fmt.Printf("Usage of our Program: \n")
		fmt.Printf("./bridge -mode [serve|init] -cnf mynode.env\n")
	}
	flag.Parse()
	logrus.Printf("mode %v path %v", mode, path)
	if mode == "serve" {
		err := NewNode(path)
		if err != nil {
			logrus.Fatalf("NewNode %v", err)
			panic(err)
		}
	} else {
		err := nodeInit(path)
		if err != nil {
			logrus.Fatalf("nodeInit %v", err)
			panic(err)
		}
	}

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

func (n Node) newBLSNode(num int) *modelBLS.Node {
	n.P2PPubSub = new(libp2p_pubsub.Libp2pPubSub)
	n.P2PPubSub.InitializePubSub(n.Host)
	suite := pairing.NewSuiteBn256()
	prvKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(prvKey, nil)
	publicKeys := make([]kyber.Point, 0)
	publicKeys = append(publicKeys, pubKey)
	mask, _ := sign.NewMask(suite, publicKeys, nil)
	fmt.Printf("HostId %v\n", n.Host.ID())
	fmt.Printf("Host config Id %v\n", config.Config.ID)
	return &modelBLS.Node{
		Id:           config.Config.ID,
		TimeStep:     0,
		ThresholdWit: num/2 + 1,
		ThresholdAck: num/2 + 1,
		Acks:         0,
		ConvertMsg:   &messageSigpb.Convert{},
		Comm:         n.P2PPubSub,
		History:      make([]modelBLS.MessageWithSig, 0),
		Signatures:   make([][]byte, num),
		SigMask:      mask,
		PublicKeys:   publicKeys,
		PrivateKey:   prvKey,
		Suite:        suite,
	}
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

func nodeInit(path string) (err error) {
	logrus.Print("nodeInit START")
	pub, err := common2.CreateBN256Key()
	if err != nil {
		return
	}
	logrus.Printf("pubkey %v", pub)
	//common.ReadPointFromFile("")
	common2.CreateRSAKey()
	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	logrus.Printf("keyfile %v port %v", config.Config.KEY_FILE, config.Config.P2P_PORT)
	h, err := libp2p.NewHost(context.Background(), 0, config.Config.KEY_FILE, config.Config.P2P_PORT)
	if err != nil {
		return
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, "keys/peer.env")
	c1, c2, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Printf("nodelist 1 %v", common.HexToAddress(config.Config.ECDSA_KEY_1))

	err = common2.RegisterNode(c1, common.HexToAddress(config.Config.NODELIST_NETWORK1), common.HexToAddress(config.Config.ECDSA_KEY_1), []byte(nodeURL), []byte(pub))
	if err != nil {
		return
	}

	logrus.Printf("nodelist 2 %v", common.HexToAddress(config.Config.NODELIST_NETWORK2))
	err = common2.RegisterNode(c2, common.HexToAddress(config.Config.NODELIST_NETWORK2), common.HexToAddress(config.Config.ECDSA_KEY_2), []byte(nodeURL), []byte(pub))
	if err != nil {
		return
	}

	return
}
