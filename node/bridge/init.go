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
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-flow-metrics"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
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

	for _, chain := range config.Bridge.Chains {
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
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, keyFile, portFromFile, ipFromFile)
	if err != nil {
		logrus.Fatal(err)
	}

	n.Dht, err = n.InitDHT(config.Bridge.BootstrapAddrs)
	if err != nil {
		return err
	}

	if n.gsnClient, err = gsn.NewClient(n.Ctx, n.Host, n, config.Bridge.TickerInterval); err != nil {
		logrus.Fatal(err)
	}

	NodeIdAddressFromFile := common.BytesToAddress([]byte(n.Host.ID()))
	sentry.AddTags(map[string]string{
		field.PeerId:         n.Host.ID().Pretty(),
		field.NodeAddress:    NodeIdAddressFromFile.Hex(),
		field.NodeRendezvous: config.Bridge.Rendezvous,
	})

	var cl *Client
	for _, client := range n.Clients {
		if client.EthClient != nil {
			cl = &client
			break
		}
	}
	if cl == nil {
		return fmt.Errorf("node  client  not initialized")
	}

	res, err := cl.NodeList.NodeExists(NodeIdAddressFromFile)
	if err != nil {

		logrus.Error(fmt.Errorf("can not check node existence on error: %w", err))
	}
	if res {
		nodeFromContract, err := cl.NodeList.GetNode(NodeIdAddressFromFile)
		if err != nil {
			return err
		}
		logrus.Infof("PORT %d", portFromFile)

		logrus.Infof("Node address: %x nodeAddress from contract: %x", common.BytesToAddress([]byte(n.Host.ID())), nodeFromContract.NodeIdAddress)

		if nodeFromContract.NodeIdAddress != common.BytesToAddress([]byte(n.Host.ID())) {
			logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", nodeFromContract.NodeIdAddress, common.BytesToAddress([]byte(n.Host.ID())))
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
				big.NewInt(int64(client.ChainCfg.Id)))
			if errors.Is(err, ErrGetEthClient) {
				logrus.Warnf("ListenNodeOracleRequest on not reachable network [%d]",
					client.ChainCfg.Id)
				continue
			} else if errors.Is(err, ErrContextDone) {
				logrus.Info(err)
				return nil
			} else if err != nil {
				return fmt.Errorf("stop listen for node oracle request chainId [%s] on error: %w", chainIdString, err)
			} else {
				logrus.Infof("ListenNodeOracleRequest on chain [%d]",
					client.ChainCfg.Id)
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

func (n *Node) GetNodeClientOrRecreate(chainId *big.Int) (Client, bool, error) {

	n.cMx.Lock()
	defer n.cMx.Unlock()
	clientRecreated := false
	client, ok := n.Clients[chainId.String()]
	if !ok {
		return Client{}, false, errors.New("eth client for chain ID not found")
	} else if client.EthClient == nil {
		var err error
		client.EthClient, client.currentUrl, err = config.Bridge.Chains.GetChainCfg(uint(chainId.Uint64())).GetEthClient("")
		if err != nil {
			return client, false, ErrGetEthClient
		} else if err = client.RecreateContractsAndFilters(); err != nil {

			return client, false, ErrGetEthClient
		} else {
			clientRecreated = true
			n.Clients[chainId.String()] = client
		}
	}
	ctx, cancel := context.WithTimeout(n.Ctx, 3*time.Second)
	defer cancel()

	netChainId, err := client.EthClient.ChainID(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			field.CainId: chainId,
		}).Error(fmt.Errorf("recreate network client on error: %w", err))

		client, err = NewClient(client.ChainCfg, client.currentUrl)
		if errors.Is(err, ErrGetEthClient) {

			return client, false, ErrGetEthClient
		} else if err != nil {

			err = fmt.Errorf("can not create client on error:%w", err)
			return Client{}, false, err
		} else {
			// replace network client in clients map
			clientRecreated = true
			n.Clients[chainId.String()] = client
			logrus.Infof("client for chain[%s] recrated", chainId.String())
			return client, clientRecreated, nil
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

func (n *Node) IsClientReady(chainId *big.Int) bool {
	n.cMx.Lock()
	defer n.cMx.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, ok := n.Clients[chainId.String()]
	if !ok {
		logrus.Errorf("eth client for chain[%s] not found", chainId.String())
		return false
	} else if client.EthClient == nil {
		logrus.Errorf("eth client for chain[%s] not initialized", chainId.String())
		return false
	} else if _, err := client.EthClient.ChainID(ctx); err != nil {
		logrus.Error(fmt.Errorf("network with chain[%s] is not reachable on error: %w", chainId.String(), err))
		return false
	} else {
		return true
	}

}

func (n Node) GetNodeClient(chainId *big.Int) (Client, error) {
	n.cMx.Lock()
	defer n.cMx.Unlock()

	client, ok := n.Clients[chainId.String()]
	if !ok {
		return Client{}, errors.New("eth client for chain ID not found")
	} else if client.EthClient == nil {
		return client, ErrGetEthClient
	}

	return client, nil
}

func NewNodeWithClients(ctx context.Context) (n *Node, err error) {
	n = &Node{
		Node: base.Node{
			Ctx: ctx,
		},
		nonceMx:        new(sync.Mutex),
		cMx:            new(sync.Mutex),
		Clients:        make(map[string]Client, len(config.Bridge.Chains)),
		uptimeRegistry: new(flow.MeterRegistry),
	}
	logrus.Print(len(config.Bridge.Chains), " chains Length")

	for _, chain := range config.Bridge.Chains {
		key := fmt.Sprintf("%d", chain.Id)
		logrus.Print("CHAIN ", chain, "chain")
		client, err := NewClient(chain, "")
		if err != nil {

			logrus.Error(fmt.Errorf("init chain[%d] node client error: %w", chain.Id, err))
			// return nil, fmt.Errorf("init chain[%d] node client error: %w", chain.Id, err)
		} else if reflect.DeepEqual(client, ethclient.Client{}) {

			logrus.Error(fmt.Errorf("init chain [%d] client failed", chain.Id))
			// return nil, fmt.Errorf("init chain [%d] client failed", chain.Id)
		} else if _, ok := n.Clients[key]; ok {

			logrus.Error(fmt.Errorf("init duplicate  chain[%d] node client chainId:[%s] error %w", chain.Id, client.ChainCfg.ChainId, err))
			// return nil, fmt.Errorf("init duplicate  chain[%d] node client chainId:[%s] error %w", chain.Id, client.ChainCfg.ChainId, err)
		}

		n.Clients[key] = client
	}
	return
}
