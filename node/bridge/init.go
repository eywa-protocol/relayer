package bridge

import (
	"context"
	"errors"
	"fmt"
	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	"github.com/digiu-ai/p2p-bridge/runa"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
)

type BsnConfig struct {
	BootstrapAddrs []string `yaml:"bootstrap-addrs"`
}

const bsnConfigFile = "bsn.yaml"

func loadConfigSetKeysAndCheck(path string) {
	if err := loadConfig(path); err != nil {
		logrus.Fatal(err)
	}
	loadKeys()
	if err := checkKeys(); err != nil {
		logrus.Fatal(err)
	}
	checkBalances()
}

func loadConfig(path string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Tracef("started in directory %s", dir)
	err = config.LoadConfigAndArgs(path)
	return
}

func loadKeys() {
	config.Config.ECDSA_KEY_1 = os.Getenv("ECDSA_KEY_1")
	config.Config.ECDSA_KEY_2 = os.Getenv("ECDSA_KEY_2")
	config.Config.ECDSA_KEY_3 = os.Getenv("ECDSA_KEY_3")
}

func checkKeys() (err error) {
	if config.Config.ECDSA_KEY_1 == "" || config.Config.ECDSA_KEY_2 == "" || config.Config.ECDSA_KEY_3 == "" {
		return errors.New("you need key to start node")
	}
	return nil
}

func checkBalances() {

	c1, c2, c3, err := getEthClients()
	if err != nil {
		logrus.Fatal(err)

	}

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
		logrus.Fatalf("you need balance on your wallets 1: %d 2: %d 3: %d to start node", balance1, balance2, balance3)
	}
}

func InitNode(path, name, keysPath string) (err error) {

	loadConfigSetKeysAndCheck(path)

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

func NewNode(path, keysPath, name, rendezvous string) (err error) {
	loadConfigSetKeysAndCheck(path)
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
		logrus.Fatalf("Can't obtain port on error: %v", err)
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
		runa.Host(n.Host, cancel)
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
		EthClient: c,
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
	chainIdToCompare, err := client.EthClient.ChainID(context.Background())
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

	bsnConfigData, err := ioutil.ReadFile(bsnConfigFile)
	if err != nil {
		err = fmt.Errorf("can not resd bootstrap nodes config %s on error: %w", bsnConfigFile, err)
		logrus.Error(err)
		return nil, err
	}

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
