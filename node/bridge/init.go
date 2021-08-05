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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-flow-metrics"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
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
		client, url, err := chain.GetEthClient("")
		if err != nil {
			return fmt.Errorf("%w", err)
		} else {
			logrus.Tracef("chain[%s] client connected to url: %s", chain.ChainId.String(), url)
		}
		if err := common2.RegisterNode(client, chain.EcdsaKey, chain.NodeListAddress, h.ID(), pub); err != nil {
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
	sentry.AddTags(map[string]string{
		field.PeerId:         hostFromFile.ID().Pretty(),
		field.NodeAddress:    NodeIdAddressFromFile.Hex(),
		field.NodeRendezvous: config.App.Rendezvous,
	})

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
				client.ChainCfg.ChainId)
			if errors.Is(err, ErrContextDone) {
				logrus.Info(err)
				return nil
			} else if err != nil {
				return fmt.Errorf("stop listen for node oracle request chainId [%s] on error: %w", chainIdString, err)
			}
		}

		wg.Add(1)
		go n.UptimeSchedule(wg)

		logrus.Info("bridge started")
		runa.Host(n.Host, cancel, wg)
		return nil
	}
	return
}

func NewClient(chain *config.Chain, skipUrl string) (client Client, err error) {

	c, url, err := chain.GetEthClient(skipUrl)
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
		currentUrl:       url,
	}, nil
}

func (n *Node) GetNodeClientOrRecreate(chainId *big.Int) (Client, bool, error) {
	n.cMx.Lock()
	clientRecreated := false
	client, ok := n.Clients[chainId.String()]
	if !ok {
		n.cMx.Unlock()
		return Client{}, false, errors.New("eth client for chain ID not found")
	}
	n.cMx.Unlock()

	netChainId, err := client.EthClient.ChainID(n.Ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			field.CainId: chainId,
		}).Error(fmt.Errorf("recreate network client on error: %w", err))

		client, err = NewClient(client.ChainCfg, client.currentUrl)
		if err != nil {
			err = fmt.Errorf("can not create client on error:%w", err)
			return Client{}, false, err
		} else {
			// replace network client in clients map
			clientRecreated = true
			n.cMx.Lock()
			n.Clients[chainId.String()] = client
			n.cMx.Unlock()
		}
	}

	if netChainId != nil && netChainId.Cmp(chainId) != 0 {
		err = errors.New("chain id not match to network")
		logrus.WithFields(logrus.Fields{
			field.CainId:         chainId,
			field.NetworkChainId: netChainId,
		}).Error(err)
		return Client{}, false, err
	}

	return client, clientRecreated, nil
}

func (n Node) GetNodeClient(chainId *big.Int) (Client, error) {
	n.cMx.Lock()
	defer n.cMx.Unlock()

	client, ok := n.Clients[chainId.String()]
	if !ok {
		return Client{}, errors.New("eth client for chain ID not found")
	}

	return client, nil
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
		Ctx:            ctx,
		nonceMx:        new(sync.Mutex),
		cMx:            new(sync.Mutex),
		Clients:        make(map[string]Client, len(config.App.Chains)),
		uptimeRegistry: new(flow.MeterRegistry),
	}
	logrus.Print(len(config.App.Chains), " chains Length")
	for _, chain := range config.App.Chains {
		logrus.Print("CHAIN ", chain, "chain")
		client, err := NewClient(chain, "")
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
