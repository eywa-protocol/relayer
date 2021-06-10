package node

import (
	"context"
	"errors"
	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func loadNodeConfig(path string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Tracef("started in directory %s", dir)
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

	// TODO: SCALED_NUM приходит ""
	strNum := strings.TrimPrefix(os.Getenv("SCALED_NUM"), "p2p-bridge_node_")
	nodeHostId, _ := strconv.Atoi(strNum)
	if common2.FileExists("keys/scaled-num-peer.log") {
		nodeHostIdB, err := ioutil.ReadFile("keys/scaled-num-peer.log")
		if err != nil {
			panic(err)
		}
		nodeHostId, _ = strconv.Atoi(string(nodeHostIdB))
	} else {

		err = ioutil.WriteFile("keys/scaled-num-peer.log", []byte(strconv.Itoa(nodeHostId)), 0644)
		if err != nil {
			panic(err)
		}
	}

	c1, c2, c3, err := getEthClients()
	if err != nil {
		logrus.Fatal(err)

	}

	var getRandomKeyForTestIfNoFunds func()
	getRandomKeyForTestIfNoFunds = func() {
		config.Config.ECDSA_KEY_1 = keys[nodeHostId-1]
		config.Config.ECDSA_KEY_2 = keys2[nodeHostId-1]
		balance1, err := c1.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_1), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		balance2, err := c2.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_2), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		balance3, err := c3.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_3), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		if balance1 == big.NewInt(0) || balance2 == big.NewInt(0) || balance3 == big.NewInt(0) {
			logrus.Errorf("you need balance on your wallets 1: %d 2: %d to start node", balance1, balance2)
			// if nodeHostId == 0 || nodeHostId > len(keys)-1 {
			rand.Seed(time.Now().UnixNano())
			nodeHostId = rand.Intn(len(keys))
			// }
			getRandomKeyForTestIfNoFunds()
		}

	}
	config.Config.ECDSA_KEY_1 = keys[nodeHostId]
	config.Config.ECDSA_KEY_2 = keys2[nodeHostId]

	if config.Config.ECDSA_KEY_1 == "" || config.Config.ECDSA_KEY_2 == "" {
		panic(errors.New("you need key to start node"))
	}

	logrus.Tracef("hostName %s nodeHostId %v key1 %v key2 %v", hostName, nodeHostId, config.Config.ECDSA_KEY_1, config.Config.ECDSA_KEY_2)

	return
}

func NodeInit(path, name, keysPath string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	if common2.FileExists(keysPath + name + "-ecdsa.key") {
		return errors.New("node allready registered! ")
	}

	err = common2.GenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		panic(err)
	}

	_, _, err = common2.GenAndSaveBN256Key(keysPath, name)
	if err != nil {
		panic(err)
	}

	blsAddr, pub, err := common2.GenAndSaveBN256Key(keysPath, name)
	if err != nil {
		return
	}

	err = common2.GenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		return
	}

	logrus.Tracef("keyfile %v", keysPath+"/"+name+"-ecdsa.key")

	h, err := libp2p.NewHostFromKeyFila(context.Background(), keysPath+"/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")
	c1, c2, _, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Infof("nodelist1 blsAddress: %v", blsAddr)
	pKey1, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	err = common2.RegisterNode(c1, pKey1, common.HexToAddress(config.Config.NODELIST_NETWORK1), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network1 %v", err)
	}

	pKey2, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c2, pKey2, common.HexToAddress(config.Config.NODELIST_NETWORK2), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network2 %v", err)
	}

	return
}

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logrus.Infof("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func NewNode(path, name string, rendezvous string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	n := &Node{
		Ctx: ctx,
	}
	// n.PublicKeys = make([]kyber.Point,0)

	n.pKey, err = common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Tracef("n.Config.PORT_1 %d", config.Config.PORT_1)

	n.EthClient_1, n.EthClient_2, n.EthClient_3, err = getEthClients()
	if err != nil {
		return
	}

	key_file := "keys/" + name + "-ecdsa.key"
	bls_key_file := "keys/" + name + "-bn256.key"

	blsAddr, err := common2.BLSAddrFromKeyFile(bls_key_file)
	if err != nil {
		return
	}

	q, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	if err != nil {
		return
	}

	words := strings.Split(string(q.P2pAddress), "/")

	registeredPort, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Fatalf("Can't obtain port %d, %v", registeredPort, err)
	}
	logrus.Infof("PORT %d", registeredPort)

	registeredAddress := words[2]

	registeredPeer, err := ioutil.ReadFile("keys/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", "keys/"+name+"-peer.env", err)
	}
	logrus.Infof("Node address: %s nodeAddress from contract: %s", string(registeredPeer), string(q.P2pAddress))

	if string(registeredPeer) != string(q.P2pAddress) {
		logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", string(registeredPeer), string(q.P2pAddress))
	}
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, key_file, registeredPort, registeredAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("host id %v address %v", n.Host.ID(), n.Host.Addrs()[0])

	n.DiscoveryPeers, err = n.setDiscoveryPeers()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Printf("setDiscoveryPeers len(n.DiscoveryPeers)=%d", len(n.DiscoveryPeers))

	n.Dht, err = n.initDHT()
	if err != nil {
		return
	}

	//
	// ======== 4. AFTER CONNECTION TO BOOSTRAP NODE WE ARE DISCOVERING OTHER ========
	//

	n.P2PPubSub = n.InitializeCoomonPubSub()
	n.P2PPubSub.InitializePubSubWithTopic(n.Host, rendezvous)

	go n.DiscoverByRendezvous(rendezvous)

	n.PrivKey, n.BLSAddress, err = n.KeysFromFilesByConfigName(name)
	if err != nil {
		return
	}
	eventChan := make(chan *wrappers.BridgeOracleRequest)
	wg := &sync.WaitGroup{}
	eventChan2 := make(chan *wrappers.BridgeOracleRequest)
	wg2 := &sync.WaitGroup{}
	defer wg.Done()
	err = n.ListenNodeOracleRequest(
		eventChan,
		wg,
		config.Config.PROXY_NETWORK1,
		config.Config.NODELIST_NETWORK1,
		n.EthClient_1,
		config.Config.ECDSA_KEY_2,
		config.Config.PROXY_NETWORK2)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	err = n.ListenNodeOracleRequest(
		eventChan2,
		wg2,
		config.Config.PROXY_NETWORK2,
		config.Config.NODELIST_NETWORK2,
		n.EthClient_2,
		config.Config.ECDSA_KEY_1,
		config.Config.PROXY_NETWORK1)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	err = n.ListenNodeOracleRequest(
		eventChan2,
		wg2,
		config.Config.PROXY_NETWORK2,
		config.Config.NODELIST_NETWORK2,
		n.EthClient_3,
		config.Config.ECDSA_KEY_1,
		config.Config.PROXY_NETWORK1)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	// n.ListenReceiveRequest(n.EthClient_2, common.HexToAddress(config.Config.PROXY_NETWORK2))

	logrus.Info("bridge started")
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	// n.Server.Start(port)

	run(n.Host, cancel)
	return
}

func getEthClients() (c1 *ethclient.Client, c2 *ethclient.Client, c3 *ethclient.Client, err error) {
	logrus.Tracef("config.Config.NETWORK_RPC_1 %s", config.Config.NETWORK_RPC_1)
	c1, err = ethclient.Dial(config.Config.NETWORK_RPC_1)
	if err != nil {
		return
	}

	c2, err = ethclient.Dial(config.Config.NETWORK_RPC_2)
	if err != nil {
		return
	}
	c3, err = ethclient.Dial(config.Config.NETWORK_RPC_3)
	if err != nil {
		return
	}
	return
}

func (n Node) setDiscoveryPeers() (discoveryPeers []multiaddr.Multiaddr, err error) {
	discoveryPeers = make([]multiaddr.Multiaddr, 0)
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		peerMA, err := multiaddr.NewMultiaddr(string(node.P2pAddress[:]))
		if err != nil {
			logrus.Errorf("setDiscoveryPeers %v", err)
		}
		discoveryPeers = append(discoveryPeers, peerMA)
	}
	return
}

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {

	dht, err = libp2p.NewDHT(n.Ctx, n.Host)
	if err != nil {
		return
	}

	return
}
