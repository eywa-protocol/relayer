package node

import (
	"context"
	"errors"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/helpers"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/host"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

func LoadNodeConfig(path string) (err error) {
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
	hostName, _ := os.Hostname()

	keysList := os.Getenv("ECDSA_KEY_1")
	keys := strings.Split(keysList, ",")
	keysList2 := os.Getenv("ECDSA_KEY_2")
	keys2 := strings.Split(keysList2, ",")

	strNum := strings.TrimPrefix(hostName, "p2p-bridge_node_")
	strNum = strings.TrimRightFunc(strNum, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	nodeHostId, _ := strconv.Atoi(strNum)

	c1, c2, err := GetEthClients()
	if err != nil {
		logrus.Fatal(err)

	}

	var getRandomKeyForTestIfNoFunds func()
	getRandomKeyForTestIfNoFunds = func() {
		config.Config.ECDSA_KEY_1 = keys[nodeHostId]
		config.Config.ECDSA_KEY_2 = keys2[nodeHostId]
		balance1, err := c1.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_1), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		balance2, err := c2.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_2), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		if balance1 == big.NewInt(0) || balance2 == big.NewInt(0) {
			logrus.Errorf("you need balance on your wallets 1: %d 2: %d to start node", balance1, balance2)
			//if nodeHostId == 0 || nodeHostId > len(keys)-1 {
			rand.Seed(time.Now().UnixNano())
			nodeHostId = rand.Intn(len(keys))
			//}
			getRandomKeyForTestIfNoFunds()
		}

	}

	config.Config.ECDSA_KEY_1 = keys[nodeHostId]
	config.Config.ECDSA_KEY_2 = keys2[nodeHostId]

	if config.Config.ECDSA_KEY_1 == "" || config.Config.ECDSA_KEY_2 == "" {
		panic(errors.New("you need key to start node"))
	}

	logrus.Printf("hostName %s nodeHostId %v key1 %v key2 %v", hostName, nodeHostId, config.Config.ECDSA_KEY_1, config.Config.ECDSA_KEY_2)

	return
}

func NodeInit(path, name string) (err error) {
	logrus.Print("nodeInit START")

	err = LoadNodeConfig(path)
	if err != nil {
		return
	}

	blsAddr, pub, err := common2.CreateBN256Key(name)
	if err != nil {
		return
	}

	//logrus.Printf("pubkey %v", pub)

	err = common2.GenECDSAKey(name)
	if err != nil {
		return
	}

	logrus.Printf("keyfile %v", "keys/"+name+"-ecdsa.key")
	h, err := libp2p.NewHostFromKeyFila(context.Background(), "keys/"+name+"-ecdsa.key", 0)
	if err != nil {
		return
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, "keys/"+name+"-peer.env")
	c1, c2, err := GetEthClients()

	if err != nil {
		return
	}

	logrus.Printf("nodelist1 blsAddress: %v", blsAddr)
	pKey1, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	err = common2.RegisterNode(c1, pKey1, common.HexToAddress(config.Config.NODELIST_NETWORK1), common.HexToAddress(config.Config.ECDSA_KEY_1), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network1 %v", err)
	}

	pKey2, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c2, pKey2, common.HexToAddress(config.Config.NODELIST_NETWORK2), common.HexToAddress(config.Config.ECDSA_KEY_2), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network2 %v", err)
	}

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

func NewNode(path, name string, port int) (err error) {

	err = LoadNodeConfig(path)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx:               ctx,

	}
	//n.PublicKeys = make([]kyber.Point,0)

	n.pKey, err = common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Printf("n.Config.PORT_1 %d", config.Config.PORT_1)

	n.EthClient_1, n.EthClient_2, err = GetEthClients()
	if err != nil {
		return
	}

	key_file := "keys/" + name + "-ecdsa.key"
	bls_key_file := "keys/" + name + "-bn256.key"
	blsAddr := common2.BLSAddrFromKeyFile(bls_key_file)
	q, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	if err != nil {
		return
	}

	words := strings.Split(string(q.P2pAddress), "/")

	registeredPort, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Errorf("Caanot obtain port %d, %v", registeredPort, err)
	}
	logrus.Printf(" <<<<<<<<<<<<<<<<<<<<<<< PORT %d >>>>>>>>>>>>>>>>>>>>>>>>>>>>", registeredPort)
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, key_file, registeredPort)
	if err != nil {
		return
	}
	logrus.Printf(" <<<<<<<<<<<<<<<<<<<<<<< HOST ID %v >>>>>>>>>>>>>>>>>>>>>>>>>>>>", n.Host.ID())
	n.Dht, err = n.initDHT()
	if err != nil {
		return
	}

	n.Discovery = discovery.NewRoutingDiscovery(n.Dht)
	//discovery.Advertise(n.Ctx, n.Discovery, n.CurrentRendezvous)

	//n.P2PPubSub = n.initNewPubSub()


	n.PrivKey, n.BLSAddress, err = n.KeysFromFilesByConfigName(name)
	if err != nil {
		return
	}

	logrus.Printf("---------- ListenNodeOracleRequest -------------")
	_, err = n.ListenNodeOracleRequest()
	if err != nil {
		logrus.Errorf(err.Error())
	}
	logrus.Printf("---------- ListenReceiveRequest -------------")
	helpers.ListenReceiveRequest(n.EthClient_2, common.HexToAddress(config.Config.PROXY_NETWORK2))
	//n.ListenNodeAddedEventInFirstNetwork()

	logrus.Print("newBLSNode STARTED /////////////////////////////////")
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	if port == 0 {
		port = config.Config.PORT_1
	}
	//n.Server.Start(port)

	run(n.Host, cancel)
	return
}

func GetEthClients() (c1 *ethclient.Client, c2 *ethclient.Client, err error) {
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

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {
	var bootstrapPeers []multiaddr.Multiaddr
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		peerMA, err := multiaddr.NewMultiaddr(string(node.P2pAddress[:]))
		if err != nil {
			return nil, err
		}
		bootstrapPeers = append(bootstrapPeers, peerMA)
	}
	dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers)
	if err != nil {
		return
	}
	_ = dht.Bootstrap(n.Ctx)
	return
}
