package bridge

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"

	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	"github.com/digiu-ai/p2p-bridge/runa"
	"github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

func InitNode(name, keysPath string) (err error) {

	if common2.FileExists(keysPath + name + "-ecdsa.key") {
		return errors.New("node already registered! ")
	}

	err = common2.GenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		panic(err)
	}

	_, pub, err := common2.GenAndSaveBN256Key(keysPath, name)
	if err != nil {
		return
	}

	logrus.Tracef("keyfile %v", keysPath+"/"+name+"-ecdsa.key")

	h, err := libp2p.NewHostFromKeyFila(context.Background(), keysPath+"/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	_ = libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")

	for _, chain := range config.App.Chains {
		client, err := chain.GetEthClient()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		if err := common2.RegisterNode(client, chain.EcdsaKey, chain.NodeListAddress, h.ID(), []byte(pub)); err != nil {
			return fmt.Errorf("register node on chain [%d] error: %w ", chain.Id, err)
		}
	}

	return
}

func NewNode(name, keysPath, rendezvous string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	n, err := NewNodeWithClients(ctx)
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", keysPath+"/"+name+"-peer.env", err)
	}

	keyFile := keysPath + "/" + name + "-ecdsa.key"

	peerStringFromFile, err := ioutil.ReadFile(keysPath + "/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", keysPath+"/"+name+"-peer.env", err)
	}

	words := strings.Split(string(peerStringFromFile), "/")

	portFromFile, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Fatalf("Can't obtain port on error: %v", err)
	}

	ipFromFile := words[2]
	logrus.Info("IP ADDRESS", ipFromFile)
	hostFromFile, err := libp2p.NewHostFromKeyFila(n.Ctx, keyFile, portFromFile, ipFromFile)
	if err != nil {
		logrus.Fatal(err)
	}

	NodeIdAddressFromFile := common.BytesToAddress([]byte(hostFromFile.ID()))
	c1, ok := n.Clients[config.App.Chains[0].ChainId.String()]
	if !ok {
		return fmt.Errorf("node  client 0 not initialized")
	}
	res, err := c1.NodeList.NodeExists(NodeIdAddressFromFile)

	if res {
		nodeFromContract, err := c1.NodeList.GetNode(NodeIdAddressFromFile)
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

		n.P2PPubSub = n.InitializeCommonPubSub()
		n.P2PPubSub.InitializePubSubWithTopic(n.Host, rendezvous)

		wg := &sync.WaitGroup{}

		wg.Add(1)
		go n.DiscoverByRendezvous(wg, rendezvous)

		n.PrivKey, err = n.KeysFromFilesByConfigName(name)
		if err != nil {
			return err
		}
		eventChan := make(chan *wrappers.BridgeOracleRequest)

		for chainIdString, client := range n.Clients {
			err = n.ListenNodeOracleRequest(
				eventChan,
				wg,
				client)
			if errors.Is(err, ErrContextDone) {
				logrus.Info(err)
				return nil
			} else if err != nil {
				return fmt.Errorf("stop listen for node oracle request chainId [%s] on error: %w", chainIdString, err)
			}
		}

		logrus.Info("bridge started")
		runa.Host(n.Host, cancel, wg)
		return nil
	}
	return
}

func NewClient(chain *config.Chain) (client Client, err error) {

	c, err := chain.GetEthClient()
	if err != nil {
		err = fmt.Errorf("get eth client error: %w", err)
		return
	}
	chain.ChainId, err = c.ChainID(context.Background())
	if err != nil {
		err = fmt.Errorf("get chain id error: %w", err)
		return
	}

	bridge, err := wrappers.NewBridge(chain.BridgeAddress, c)
	if err != nil {
		err = fmt.Errorf("init bridge [%s] error: %w", chain.BridgeAddress, err)
		return
	}

	bridgeFilterer, err := wrappers.NewBridgeFilterer(chain.BridgeAddress, c)
	if err != nil {
		err = fmt.Errorf("init bridge filter [%s] error: %w", chain.BridgeAddress, err)
		return
	}
	nodeList, err := wrappers.NewNodeList(chain.NodeListAddress, c)
	if err != nil {
		err = fmt.Errorf("init nodelist [%s] error: %w", chain.BridgeAddress, err)
		return
	}

	nodeListFilterer, err := wrappers.NewNodeListFilterer(chain.NodeListAddress, c)
	if err != nil {
		err = fmt.Errorf("init nodelist filter [%s] error: %w", chain.BridgeAddress, err)
		return
	}

	txOpts := common2.CustomAuth(c, chain.EcdsaKey)

	return Client{
		EthClient: c,
		ChainCfg:  chain,
		EcdsaKey:  chain.EcdsaKey,
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
	}, nil
}

func (n Node) GetNodeClientByChainId(chainIdFromClient *big.Int) (Client, error) {

	if client, ok := n.Clients[chainIdFromClient.String()]; !ok {
		return Client{}, fmt.Errorf("not found node client for chainId %s", chainIdFromClient.String())
	} else {
		return client, nil
	}
}

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {

	bootstrapPeers := make([]multiaddr.Multiaddr, 0, len(config.App.BootstrapAddrs))

	for _, addr := range config.App.BootstrapAddrs {
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

func NewNodeWithClients(ctx context.Context) (n *Node, err error) {
	n = &Node{
		Ctx:     ctx,
		nonceMx: new(sync.Mutex),
		Clients: make(map[string]Client, len(config.App.Chains)),
	}
	logrus.Print(len(config.App.Chains), " chains Length")
	for _, chain := range config.App.Chains {
		logrus.Print("CHAIN ", chain, "chain")
		client, err := NewClient(chain)
		if err != nil {
			return nil, fmt.Errorf("init chain[%d] node client error: %w", chain.Id, err)
		}
		if reflect.DeepEqual(client, ethclient.Client{}) {
			return nil, fmt.Errorf("init chain [%d] client failed", chain.Id)
		}
		if _, ok := n.Clients[client.ChainCfg.ChainId.String()]; ok {
			return nil, fmt.Errorf("init duplicate  chain[%d] node client chainId:[%s] error %w", chain.Id, client.ChainCfg.ChainId, err)
		}
		n.Clients[client.ChainCfg.ChainId.String()] = client
	}
	return
}
