package bridge

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"io/ioutil"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/cmd/utils"
	_config "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	_genesis "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/genesis"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/ledger"

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

	common2.MakeKeyDir(keysPath)

	_, err = common2.GenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		return
	}

	_, err = common2.GenAndSaveBlsKey(keysPath, name)
	if err != nil {
		return
	}

	privKey, err := common2.GetOrGenAndSaveSecp256k1Key(keysPath, name)
	if err != nil {
		return
	}

	logrus.Infoln("Generated address:")
	fmt.Println(common2.AddressFromSecp256k1PrivKey(privKey))
	logrus.Infoln("Please transfer the collateral there and restart me with -register flag.")
	return
}

func RegisterNode(name, keysPath string) (err error) {

	signerKey, err := common2.LoadSecp256k1Key(keysPath, name)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n, err := NewNodeWithClients(ctx, signerKey)
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", keysPath+"/"+name+"-peer.env", err)
	}

	// Secp256k1 node key
	// nodeKey, err := common2.GetOrGenAndSaveSecp256k1Key(keysPath, name+"-signer")

	pub, err := common2.LoadBlsPublicKey(keysPath, name)
	if err != nil {
		return
	}

	n.Host, err = libp2p.NewHostFromKeyFila(context.Background(), keysPath+"/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	_ = libp2p.WriteHostAddrToConfig(n.Host, keysPath+"/"+name+"-peer.env")

	if config.Bridge.UseGsn {

		n.Dht, err = n.InitDHT(config.Bridge.BootstrapAddrs)
		if err != nil {
			return fmt.Errorf("can not init DHT on error: %w ", err)
		}

		n.gsnClient, err = gsn.NewClient(ctx, n.Host, n, 10*time.Second)
		if err != nil {
			return fmt.Errorf("can not init gsn client on error: %w ", err)
		}
		err = n.gsnClient.WaitForDiscoveryGsn(60 * time.Second)
		if err != nil {
			return fmt.Errorf("can not discover gsn node on error: %w ", err)
		}
	}

	for _, client := range n.Clients {
		id, relayerPool, err := client.RegisterNode(n.gsnClient, signerKey, n.Host.ID(), string(pub))
		if err != nil {
			return fmt.Errorf("register node on chain [%d] error: %w ", client.ChainCfg.Id, err)
		}
		logrus.Infof("New RelayerPool created with #%d at %s.", id, relayerPool)
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
	signerKey, err := common2.LoadSecp256k1Key(keysPath, name)
	if err != nil {

		return err
	}

	n, err := NewNodeWithClients(ctx, signerKey)
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

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	sentry.AddTags(map[string]string{
		field.PeerId:         n.Host.ID().Pretty(),
		field.NodeAddress:    nodeIdAddress.Hex(),
		field.NodeRendezvous: config.Bridge.Rendezvous,
	})

	c1, ok := n.Clients[config.Bridge.Chains[0].ChainId.String()]
	if !ok {
		return fmt.Errorf("node  client 0 not initialized")
	}

	if res, err := c1.NodeRegistry.NodeExists(nodeIdAddress); err != nil {
		logrus.Fatal(fmt.Errorf("failed to check node existent on error: %w", err))
		return err
	} else if res {
		logrus.Infof("PORT %d", portFromFile)

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

		err := n.StartEpoch(c1, nodeIdAddress, rendezvous)
		if err != nil {
			return err
		}

		ledger.DefLedger, err = n.initLedger()
		if err != nil {
			logrus.Errorf("initLedger %s", err)
			return err
		}
		defer ledger.DefLedger.Close()

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
	} else {
		logrus.Warnf("node not registered")
		return nil
	}
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

		client, err = NewClient(client.ChainCfg, client.currentUrl, n.signerKey)
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

func NewNodeWithClients(ctx context.Context, signerKey *ecdsa.PrivateKey) (n *Node, err error) {
	n = &Node{
		Node: base.Node{
			Ctx: ctx,
		},
		nonceMx:        new(sync.Mutex),
		cMx:            new(sync.Mutex),
		Clients:        make(map[string]Client, len(config.Bridge.Chains)),
		signerKey:      signerKey,
		uptimeRegistry: new(flow.MeterRegistry),
	}

	logrus.Print(len(config.Bridge.Chains), " chains Length")
	for _, chain := range config.Bridge.Chains {
		logrus.Print("CHAIN ", chain, "chain")
		client, err := NewClient(chain, "", signerKey)
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

func (n *Node) initLedger() (*ledger.Ledger, error) {
	// TODO init events here
	//events.Init() //Init event hub

	var err error
	dbDir := utils.GetStoreDirPath("leveldb", _config.DefConfig.P2PNode.NetworkName)
	ledger.DefLedger, err = ledger.NewLedger(dbDir)
	if err != nil {
		return nil, fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers := []bls.PublicKey{n.EpochPublicKey}
	genesisBlock, err := _genesis.BuildGenesisBlock(bookKeepers)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	logrus.Infof("Current ChainId: %d", genesisBlock.Header.ChainID)
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("Init ledger error:%s", err)
	}

	logrus.Infof("Ledger init success")
	return ledger.DefLedger, nil
}
