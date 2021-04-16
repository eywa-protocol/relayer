package main

import (
	"context"
	"crypto/ecdsa"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/dht"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/linkpoolio/bridges"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"os"
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
}

type addrList []multiaddr.Multiaddr

func NewNode() (err error) {

	dir, err := os.Getwd()
	if err != nil {
		logrus.Warn(err)
	}
	logrus.Printf("started in directory %s", dir)
	cnf, err := config.LoadConfigAndArgs()
	if err != nil {
		logrus.Fatal(err)
	}

	n := &Node{
		Config:            *cnf,
		Ctx:               context.Background(),
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

	go func() {
		err = dht.NewDHTBootPeer(config.Config.KEY_FILE, config.Config.P2P_PORT)
		if err != nil {
			return
		}
	}()
	n.Server.Start(config.Config.PORT_1)
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

	ad3, err := common2.NewDBridge(n.EthClient_1, "SetMockPoolTestRequest", "test", common2.SetMockPoolTestRequest)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	ad4, err := common2.NewDBridge(n.EthClient_1, "ChainlinkData", "control", common2.ChainlinkData)

	if err != nil {
		logrus.Fatal(err)
		return
	}

	bridgesList = append(bridgesList, ad)
	bridgesList = append(bridgesList, ad2)
	bridgesList = append(bridgesList, ad3)
	bridgesList = append(bridgesList, ad4)
	srv = bridges.NewServer(bridgesList...)
	return
}

func main() {
	err := NewNode()
	if err != nil {
		logrus.Errorf(" NewNode %v", err)
	}
}
