package main

import (
	"context"
	"crypto/ecdsa"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/linkpoolio/bridges"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type Node struct {
	Config            config.AppConfig
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
	Comm              model.CommunicationInterface
}

type addrList []multiaddr.Multiaddr

func NewNode() (err error) {

	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Printf("started in directory %s !", dir)
	cnf, err := config.LoadConfigAndArgs()
	if err != nil {
		logrus.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Config:            *cnf,
		Ctx:               ctx,
		CurrentRendezvous: "FirstRun",
		//BRIDGE_1_ADDRESS:          common.HexToAddress(os.Getenv("BRIDGE_1_ADDRESS")),
		//ORACLE_1_ADDRESS: common.HexToAddress(os.Getenv("ORACLE_1_ADDRESS")),
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

	n.Dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers)

	n.Service = libp2p.NewService(n.Host, protocol.ID("network-start"))
	err = n.Service.SetupRPC()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	go libp2p.Discover(n.Ctx, n.Host, n.Dht, "test", config.Config.TickerInterval)
	go n.Service.StartMessaging(ctx, config.Config.TickerInterval)
	n.Server.Start(config.Config.PORT_1)
	run(n.Host, cancel)
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

	// ad3, err := common2.NewDBridge(n.EthClient_1, "SetMockPoolTestRequest", "test", common2.SetMockPoolTestRequest)
	// if err != nil {
	// 	logrus.Fatal(err)
	// 	return
	// }


	// ad4, err := common2.NewDBridge(n.EthClient_1, "ChainlinkData", "control", common2.ChainlinkData)

	if err != nil {
		logrus.Fatal(err)
		return
	}

	bridgesList = append(bridgesList, ad)
	bridgesList = append(bridgesList, ad2)
	// bridgesList = append(bridgesList, ad3)
	// bridgesList = append(bridgesList, ad4)
	srv = bridges.NewServer(bridgesList...)
	return
}

func main() {
	err := NewNode()
	if err != nil {
		logrus.Fatalf(" NewNode %v", err)
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

func (n Node) startBroadcast() {
	n.Comm.Broadcast([]byte("weqweqwe"))
}
