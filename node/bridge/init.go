package bridge

import (
	"context"
	"errors"
	"fmt"
	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
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
		return errors.New("node allready registered! ")
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
	_ = libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")
	c1, c2, c3, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Infof("nodelist1 blsAddress: %v", blsAddr)
	pKey1, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	err = common2.RegisterNode(c1, pKey1, common.HexToAddress(config.Config.NODELIST_NETWORK1), h.ID(), []byte(pub))
	if err != nil {
		logrus.Errorf("error registaring node in network1 %v", err)
	}

	pKey2, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c2, pKey2, common.HexToAddress(config.Config.NODELIST_NETWORK2), h.ID(), []byte(pub))
	if err != nil {
		logrus.Fatalf("error registaring node in network2 %v", err)

	}

	pKey3, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_3)
	if err != nil {
		logrus.Errorf("ToECDSAFromHex %v", err)
		return
	}

	err = common2.RegisterNode(c3, pKey3, common.HexToAddress(config.Config.NODELIST_NETWORK3), h.ID(), []byte(pub))
	if err != nil {
		logrus.Errorf("error registaring node in network3 %v", err)
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

	logrus.Tracef("n.Config.PORT_1 %d", config.Config.PORT_1)

	c1, c2, c3, err := getEthClients()
	if err != nil {
		return
	}

	n.Client1, err = setNodeEthClient(c1, config.Config.BRIDGE_NETWORK1, config.Config.NODELIST_NETWORK1, config.Config.ECDSA_KEY_1)
	n.Client2, err = setNodeEthClient(c2, config.Config.BRIDGE_NETWORK2, config.Config.NODELIST_NETWORK2, config.Config.ECDSA_KEY_2)
	n.Client3, err = setNodeEthClient(c3, config.Config.BRIDGE_NETWORK3, config.Config.NODELIST_NETWORK3, config.Config.ECDSA_KEY_3)

	key_file := "keys/" + name + "-ecdsa.key"

	peerStringFromFile, err := ioutil.ReadFile("keys/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", "keys/"+name+"-peer.env", err)
	}

	words := strings.Split(string(peerStringFromFile), "/")

	portFromFile, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Fatalf("Can't obtain port", err)
	}

	ipFromFile := words[2]
	logrus.Info("IP ADDRESS", ipFromFile)
	hostFromFile, err := libp2p.NewHostFromKeyFila(n.Ctx, key_file, portFromFile, ipFromFile)
	if err != nil {
		logrus.Fatal(err)
	}

	NodeidAddressFromFile := common.BytesToAddress([]byte(hostFromFile.ID()))
	res, err := n.Client1.NodeList.NodeExists(NodeidAddressFromFile)
	if res {

		nodeFromContract, err := n.Client1.NodeList.GetNode(NodeidAddressFromFile)
		if err != nil {
			return err
		}
		logrus.Infof("PORT %d", portFromFile)

		logrus.Infof("Node address: %x nodeAddress from contract: %x", common.BytesToAddress([]byte(hostFromFile.ID())), nodeFromContract.NodeIdAddress)

		if nodeFromContract.NodeIdAddress != common.BytesToAddress([]byte(hostFromFile.ID())) {
			logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", nodeFromContract.NodeIdAddress, common.BytesToAddress([]byte(hostFromFile.ID())))
		}

		n.Host = hostFromFile

		n.Dht, err = n.initDHT()
		if err != nil {
			return err
		}

		//
		// ======== 4. AFTER CONNECTION TO BOOSTRAP NODE WE ARE DISCOVERING OTHER ========
		//

		n.P2PPubSub = n.InitializeCoomonPubSub()
		n.P2PPubSub.InitializePubSubWithTopic(n.Host, rendezvous)

		go n.DiscoverByRendezvous(rendezvous)

		n.PrivKey, err = n.KeysFromFilesByConfigName(name)
		if err != nil {
			return err
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
		return nil
	}
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

func setNodeEthClient(c *ethclient.Client, bridgeAddress, nodeListAddress, ecdsa_key string) (client Client, err error) {
	pKey, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

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

	txOpts := common2.CustomAuth(c, pKey)

	client = Client{
		ethClient: c,
		ECDSA_KEY: ecdsa_key,
		Bridge: wrappers.BridgeSession{
			Contract:     bridge,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
		},
		NodeList: wrappers.NodeListSession{
			Contract:     nodeList,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
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

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {

	dht, err = libp2p.NewDHT(n.Ctx, n.Host)
	if err != nil {
		return
	}

	return
}
