package node

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
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

	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	"github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-yaml/yaml"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

type BsnConfig struct {
	BootstrapAddrs []string `yaml:"bootstrap-addrs"`
}

//go:embed bsn.yaml
var bsnConfigData []byte

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
	keysList3 := os.Getenv("ECDSA_KEY_3")
	keys3 := strings.Split(keysList3, ",")

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
		config.Config.ECDSA_KEY_3 = keys3[nodeHostId-1]
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
			rand.Seed(time.Now().UnixNano())
			nodeHostId = rand.Intn(len(keys))
			getRandomKeyForTestIfNoFunds()
		}

	}
	config.Config.ECDSA_KEY_1 = keys[nodeHostId]
	config.Config.ECDSA_KEY_2 = keys2[nodeHostId]
	config.Config.ECDSA_KEY_3 = keys3[nodeHostId]

	if config.Config.ECDSA_KEY_1 == "" || config.Config.ECDSA_KEY_2 == "" || config.Config.ECDSA_KEY_3 == "" {
		panic(errors.New("you need key to start node"))
	}

	logrus.Tracef("hostName %s nodeHostId %v key1 %v key2 %v key3 %v", hostName, nodeHostId, config.Config.ECDSA_KEY_1, config.Config.ECDSA_KEY_2, config.Config.ECDSA_KEY_3)

	return
}

func InitNode(path, name, keysPath string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	if common2.FileExists(keysPath + name + "-ecdsa.key") {
		return errors.New("node already registered! ")
	}

	err = common2.GenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		panic(err)
	}

	blsAddr, pub, err := common2.GenAndSaveBN256Key(keysPath, name)
	if err != nil {
		return
	}

	logrus.Tracef("keyfile %v", keysPath+"/"+name+"-ecdsa.key")

	h, err := libp2p.NewHostFromKeyFila(context.Background(), keysPath+"/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")
	c1, c2, c3, err := getEthClients()

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
		logrus.Fatalf("error registaring node in network2 %v", err)

	}

	pKey3, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_3)
	if err != nil {
		logrus.Errorf("ToECDSAFromHex %v", err)
		return
	}

	err = common2.RegisterNode(c3, pKey3, common.HexToAddress(config.Config.NODELIST_NETWORK3), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network3 %v", err)
	}

	return
}

func BootstrapNodeInit(keysPath, name string) (err error) {

	keyFile := keysPath + "/" + name + "-rsa.key"
	if common2.FileExists(keyFile) {
		return errors.New("node already registered! ")
	}

	r := rand.New(rand.NewSource(int64(4001)))

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	pkData, err := crypto.MarshalPrivateKey(prvKey)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(keyFile, pkData, 0644)
	if err != nil {
		panic(err)
	}

	multiAddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
	if err != nil {
		panic(err)
	}

	h, err := libp2p.NewHostFromRsaKey(prvKey, multiAddr)
	if err != nil {
		panic(err)
	}

	nodeURL := libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")

	logrus.Infof("init bootstrap node: %s", nodeURL)

	return
}

func NewBootstrapNode(keysPath, name string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	keyFile := keysPath + "/" + name + "-rsa.key"

	pkData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		logrus.Fatalf("can not read private key file [%s] on error: %v", keyFile, err)
	}

	pk, err := crypto.UnmarshalPrivateKey(pkData)
	if err != nil {
		logrus.Fatalf("unmarshal private key error: %v", err)
	}

	registeredPeer, err := ioutil.ReadFile(keysPath + "/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", "keys/"+name+"-peer.env", err)
	}

	words := strings.Split(string(registeredPeer), "/")

	registeredPort, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Fatalf("Can't obtain port %d, %v", registeredPort, err)
	}
	logrus.Infof("PORT %d", registeredPort)

	registeredAddress := words[2]

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", registeredAddress, registeredPort))
	if err != nil {
		logrus.Fatalf("create multiaddr error: %v", err)
	}

	h, err := libp2p.NewHostFromRsaKey(pk, multiAddr)
	if err != nil {
		logrus.Fatal(fmt.Errorf("new bootstrap host error: %w", err))
	}

	_, err = dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	run(h, cancel)

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

func NewNode(path, keysPath, name, rendezvous string) (err error) {

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

	logrus.Tracef("n.Config.PORT_1 %d", config.Config.PORT_1)

	c1, c2, c3, err := getEthClients()
	if err != nil {
		return
	}

	n.Client1, err = setNodeEthClient(c1, config.Config.BRIDGE_NETWORK1, config.Config.NODELIST_NETWORK1, config.Config.ECDSA_KEY_1)
	n.Client2, err = setNodeEthClient(c2, config.Config.BRIDGE_NETWORK2, config.Config.NODELIST_NETWORK2, config.Config.ECDSA_KEY_2)
	n.Client3, err = setNodeEthClient(c3, config.Config.BRIDGE_NETWORK3, config.Config.NODELIST_NETWORK3, config.Config.ECDSA_KEY_3)

	keyFile := keysPath + "/" + name + "-ecdsa.key"
	blsKeyFile := keysPath + "/" + name + "-bn256.key"

	blsAddr, err := common2.BLSAddrFromKeyFile(blsKeyFile)
	if err != nil {
		return
	}

	q, err := n.Client1.NodeList.GetNode(blsAddr)
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

	registeredPeer, err := ioutil.ReadFile(keysPath + "/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", "keys/"+name+"-peer.env", err)
	}
	logrus.Infof("Node address: %s nodeAddress from contract: %s", string(registeredPeer), string(q.P2pAddress))

	if string(registeredPeer) != string(q.P2pAddress) {
		logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", string(registeredPeer), string(q.P2pAddress))
	}
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, keyFile, registeredPort, registeredAddress)
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
	defer wg.Done()

	err = n.ListenNodeOracleRequest(
		eventChan,
		wg,
		n.Client1)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	err = n.ListenNodeOracleRequest(
		eventChan,
		wg,
		n.Client2)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	err = n.ListenNodeOracleRequest(
		eventChan,
		wg,
		n.Client3)
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	logrus.Info("bridge started")
	run(n.Host, cancel)
	return
}

func getEthClients() (c1, c2, c3 *ethclient.Client, err error) {
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

func setNodeEthClient(c *ethclient.Client, bridgeAddress, nodeListAddress, ecdsaKey string) (client Client, err error) {
	bridge, err := wrappers.NewBridge(common.HexToAddress(bridgeAddress), c)
	if err != nil {
		return
	}

	bridgeFilterer, err := wrappers.NewBridgeFilterer(common.HexToAddress(bridgeAddress), c)
	if err != nil {
		return
	}
	nodeList, err := wrappers.NewNodeList(common.HexToAddress(nodeListAddress), c)
	if err != nil {
		return
	}

	nodeListFilterer, err := wrappers.NewNodeListFilterer(common.HexToAddress(nodeListAddress), c)
	if err != nil {
		return
	}

	client = Client{
		ethClient: c,
		ECDSA_KEY: ecdsaKey,
		Bridge: wrappers.BridgeSession{
			Contract:     bridge,
			CallOpts:     bind.CallOpts{},
			TransactOpts: bind.TransactOpts{},
		},
		NodeList: wrappers.NodeListSession{
			Contract:     nodeList,
			CallOpts:     bind.CallOpts{},
			TransactOpts: bind.TransactOpts{},
		},
		BridgeFilterer:   *bridgeFilterer,
		NodeListFilterer: *nodeListFilterer,
	}
	return
}

func CompareChainIds(client Client, chainId *big.Int) (equal bool, err error) {
	chainIdToCompare, err := client.ethClient.ChainID(context.Background())
	if err != nil {
		return
	}
	if chainIdToCompare.Cmp(chainId) == 0 {
		return true, nil
	}
	return false, nil
}

func (n Node) GetNodeClientByChainId(chainIdFromClient *big.Int) (client Client, err error) {

	if equals, _ := CompareChainIds(n.Client1, chainIdFromClient); equals {
		return n.Client1, nil
	}

	if equals, _ := CompareChainIds(n.Client2, chainIdFromClient); equals {
		return n.Client2, nil
	}

	if equals, _ := CompareChainIds(n.Client3, chainIdFromClient); equals {
		return n.Client3, nil
	}

	return Client{}, errors.New(fmt.Sprintf("not found rpcurl for chainID %d", chainIdFromClient))
}

func (n Node) setNodeClients(c1, c2, c3 *ethclient.Client) (err error) {

	n.Client1, err = setNodeEthClient(c1, config.Config.BRIDGE_NETWORK1, config.Config.NODELIST_NETWORK1, config.Config.ECDSA_KEY_1)
	n.Client2, err = setNodeEthClient(c2, config.Config.BRIDGE_NETWORK2, config.Config.NODELIST_NETWORK2, config.Config.ECDSA_KEY_2)
	n.Client3, err = setNodeEthClient(c3, config.Config.BRIDGE_NETWORK3, config.Config.NODELIST_NETWORK3, config.Config.ECDSA_KEY_3)

	return
}

func (n Node) setDiscoveryPeers() (discoveryPeers []multiaddr.Multiaddr, err error) {
	discoveryPeers = make([]multiaddr.Multiaddr, 0)
	nodes, err := common2.GetNodesFromContract(n.Client1.ethClient, common.HexToAddress(config.Config.NODELIST_NETWORK1))
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

	var bsnConfig BsnConfig

	err = yaml.Unmarshal(bsnConfigData, &bsnConfig)
	if err != nil {
		return nil, err
	}

	bootstrapPeers := make([]multiaddr.Multiaddr, 0, len(bsnConfig.BootstrapAddrs))

	for _, addr := range bsnConfig.BootstrapAddrs {
		logrus.Infof("add bootstrap peer: %s", addr)
		nAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		bootstrapPeers = append(bootstrapPeers, nAddr)
	}
	logrus.Infof("bootstrap peers count: %d", len(bootstrapPeers))
	dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers[:])
	if err != nil {
		return
	}

	return
}
