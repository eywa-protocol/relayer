package bridge

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/cmd/utils"
	_config "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	_genesis "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/genesis"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/ledger"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/forward"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/contracts"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

	n, err, clientsClose := NewNodeWithClients(ctx, signerKey)
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", keysPath+"/"+name+"-peer.env", err)
	}
	defer clientsClose()

	pub, err := common2.LoadBlsPublicKey(keysPath, name)
	if err != nil {
		return
	}

	n.Host, err = libp2p.NewHostFromKeyFile(context.Background(), keysPath+"/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	_ = libp2p.WriteHostAddrToConfig(n.Host, keysPath+"/"+name+"-peer.env")

	if config.Bridge.UseGsn {

		n.Dht, err = n.InitDHT(config.Bridge.BootstrapAddrs)
		if err != nil {
			return fmt.Errorf("can not init DHT on error: %w ", err)
		}

		n.gsnClient, err = gsn.NewClient(ctx, n.Host, n, config.Bridge.GsnDiscoveryInterval)
		if err != nil {
			return fmt.Errorf("can not init gsn client on error: %w ", err)
		}
		err = n.gsnClient.WaitForDiscoveryGsn(config.Bridge.GsnWaitDuration)
		if err != nil {
			return fmt.Errorf("can not discover gsn node on error: %w ", err)
		}
	}
	regChainId := n.RegChainId()
	regClient, err := n.clients.GetEthClient(regChainId)
	if err != nil {

		return err
	}

	//

	logrus.Infof("Adding Node %s it's NodeidAddress %x", n.Host.ID(), common.BytesToAddress([]byte(n.Host.ID().String())))
	fromAddress := common2.AddressFromSecp256k1PrivKey(signerKey)
	nodeIdAsAddress := common.BytesToAddress([]byte(n.Host.ID()))

	nodeRegistry := n.NodeRegistry()
	res, err := nodeRegistry.NodeExists(&bind.CallOpts{}, nodeIdAsAddress)
	if err != nil {
		err = fmt.Errorf("node not exists nodeIdAddress: %s, client.Id: %s, error: %w",
			nodeIdAsAddress.String(), regChainId.String(), err)
	}
	if res == true {
		logrus.Infof("Node %x allready exists", n.Host.ID())
		return
	}

	eywaAddress, _ := nodeRegistry.EYWA(&bind.CallOpts{})
	eywa, err := wrappers.NewERC20Permit(eywaAddress, regClient)
	if err != nil {
		return fmt.Errorf("EYWA contract error: %w", err)
	}
	fromNonce, _ := eywa.Nonces(&bind.CallOpts{}, fromAddress)
	value, _ := eywa.BalanceOf(&bind.CallOpts{}, fromAddress)

	deadline := big.NewInt(time.Now().Unix() + 100)
	const EywaPermitName = "EYWA"
	const EywaPermitVersion = "1"
	v, r, s := common2.SignErc20Permit(signerKey, EywaPermitName, EywaPermitVersion, regChainId,
		eywaAddress, fromAddress, n.NodeRegistryAddress(), value, fromNonce, deadline)

	node := wrappers.NodeRegistryNode{
		Owner:         fromAddress,
		Pool:          common.Address{},
		NodeIdAddress: nodeIdAsAddress,
		BlsPubKey:     string(pub),
		NodeId:        big.NewInt(0),
	}

	var txHash common.Hash
	if n.CanUseGsn(regChainId) && n.gsnClient != nil {
		if txHash, err = forward.NodeRegistryCreateNode(n.gsnClient, regChainId, signerKey, n.NodeRegistryAddress(), node, deadline, v, r, s); err != nil {
			err = fmt.Errorf("CreateRelayer over gsn chainId %d ERROR: %v", regChainId, err)
			logrus.Error(err)
			return err
		}
	} else {
		if nodeRegistrySession, err := n.NodeRegistrySession(signerKey); err != nil {
			return fmt.Errorf("get node registry session error: %w", err)
		} else if tx, err := nodeRegistrySession.CreateRelayer(node, deadline, v, r, s); err != nil {
			err = fmt.Errorf("CreateRelayer chainId %d ERROR: %v", regChainId, err)
			logrus.Error(err)
			return err
		} else {
			txHash = tx.Hash()
		}
	}

	receipt, err := regClient.WaitTransaction(txHash)
	if err != nil {

		return fmt.Errorf("WaitTransaction error: %w", err)
	}
	logrus.Infof("recept.Status %d", receipt.Status)

	blockNum := receipt.BlockNumber.Uint64()

	it, err := n.NodeRegistryFilterer().FilterCreatedRelayer(&bind.FilterOpts{Start: blockNum, End: &blockNum},
		[]common.Address{node.NodeIdAddress}, []*big.Int{}, []common.Address{})
	if err != nil {

		return err
	}
	defer func() {
		if err := it.Close(); err != nil {

			logrus.Error(fmt.Errorf("close registry created rellayer iterator error: %w", err))
		}
	}()

	var (
		id          *big.Int
		relayerPool *common.Address
	)
	for it.Next() {
		logrus.Info("CreatedRelayer Event", it.Event.NodeIdAddress)
		id = it.Event.NodeId
		relayerPool = &it.Event.RelayerPool
		break
	}

	logrus.Infof("New RelayerPool created with #%d at %s.", id, relayerPool)

	return nil
}

func RunNode(name, keysPath, rendezvous string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signerKey, err := common2.LoadSecp256k1Key(keysPath, name)
	if err != nil {

		return err
	}

	n, err, clientsClose := NewNodeWithClients(ctx, signerKey)
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", keysPath+"/"+name+"-peer.env", err)
	}
	defer clientsClose()

	if n.Bridge, err = contracts.NewBridge(n.clients); err != nil {

		logrus.Fatal(fmt.Errorf("init bridge contracts error: %w", err))
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
	n.Host, err = libp2p.NewHostFromKeyFile(n.Ctx, keyFile, portFromFile, ipFromFile)
	if err != nil {
		logrus.Fatal(err)
	}

	blsNodeId, err := n.getNodeBlsId()
	if err != nil {
		return err
	}
	n.BlsNodeId = blsNodeId

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

	if res, err := n.NodeRegistry().NodeExists(&bind.CallOpts{}, nodeIdAddress); err != nil {
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

		err := n.StartEpoch(nodeIdAddress, rendezvous)
		if err != nil {
			return err
		}

		// initialize watchers
		for _, chain := range config.Bridge.Chains {
			if watcher, err := eth.NewOracleRequestWatcher(chain.BridgeAddress, n.HandleOracleRequest); err != nil {
				err = fmt.Errorf("chain [%d] init watcher error: %w", chain.Id, err)

				return err
			} else {

				n.clients.AddWatcher(new(big.Int).SetUint64(chain.Id), watcher)
			}
		}

		/*		wg.Add(1)
				go n.UptimeSchedule(wg)*/

		logrus.Info("bridge started")
		runa.Host(n.Host, cancel, wg)
		return nil
	} else {
		logrus.Warnf("node not registered")
		return nil
	}
}

// func (n *Node) GetNodeClientOrRecreate(chainId *big.Int) (Client, bool, error) {
//     n.cMx.Lock()
//     clientRecreated := false
//     client, ok := n.Clients[chainId.String()]
//     if !ok {
//         n.cMx.Unlock()
//         return Client{}, false, errors.New("eth client for chain ID not found")
//     }
//     n.cMx.Unlock()
//
//     netChainId, err := client.EthClient.ChainID(n.Ctx)
//     if err != nil {
//         logrus.WithFields(logrus.Fields{
//             field.CainId: chainId,
//         }).Error(fmt.Errorf("recreate network client on error: %w", err))
//
//         client, err = NewClient(client.ChainCfg, client.currentUrl, n.signerKey)
//         if err != nil {
//             err = fmt.Errorf("can not create client on error:%w", err)
//             return Client{}, false, err
//         } else {
//             // replace network client in clients map
//             clientRecreated = true
//             n.cMx.Lock()
//             n.Clients[chainId.String()] = client
//             n.cMx.Unlock()
//         }
//     }
//
//     if netChainId != nil && netChainId.Cmp(chainId) != 0 {
//         err = errors.New("chain id not match to network")
//         logrus.WithFields(logrus.Fields{
//             field.CainId:         chainId,
//             field.NetworkChainId: netChainId,
//         }).Error(err)
//         return Client{}, false, err
//     }
//
//     return client, clientRecreated, nil
// }

func NewNodeWithClients(ctx context.Context, signerKey *ecdsa.PrivateKey) (n *Node, err error, clientsClose func()) {
	n = &Node{
		Node: base.Node{
			Ctx: ctx,
		},
		nonceMx:        new(sync.Mutex),
		signerKey:      signerKey,
		uptimeRegistry: new(flow.MeterRegistry),
	}
	chains := make(map[uint64]*config.BridgeChain, len(config.Bridge.Chains))
	clientConfigs := make(extChains.ClientConfigs, 0, len(config.Bridge.Chains))
	logrus.Print(len(config.Bridge.Chains), " chains Length")
	for _, chain := range config.Bridge.Chains {
		ethClientConfig := &eth.Config{
			Id:   chain.Id,
			Urls: chain.RpcUrls[:],
		}
		ethClientConfig.SetDefault()
		if chain.CallTimeout > 0 {
			ethClientConfig.CallTimeout = chain.CallTimeout
		}
		if chain.DialTimeout > 0 {
			ethClientConfig.DialTimeout = chain.DialTimeout
		}
		if chain.BlockTimeout > 0 {
			ethClientConfig.BlockTimeout = chain.BlockTimeout
		}

		clientConfigs = append(clientConfigs, ethClientConfig)
		chains[chain.Id] = chain
		logrus.Debug("CHAIN ", chain, "chain")
	}

	if n.clients, err = extChains.NewClients(n.Ctx, clientConfigs); err != nil {

		return nil, fmt.Errorf("init notde clients error: %w", err), nil
	} else if regChain, ok := chains[config.Bridge.RegChainId]; !ok {

		return nil, fmt.Errorf("invalid reg chain [%d]", config.Bridge.RegChainId), nil
	} else if regClient, err := n.clients.GetEthClient(new(big.Int).SetUint64(config.Bridge.RegChainId)); err != nil {

		return nil, fmt.Errorf("get reg chain client error: %w", err), nil
	} else if n.Registry, err = contracts.NewRegistry(regChain.NodeRegistryAddress, regClient); err != nil {

		return nil, fmt.Errorf("init node registry error: %w", err), nil
	} else if n.Forwarder, err = contracts.NewForwarder(n.clients); err != nil {

		return nil, fmt.Errorf("init node forwarder error: %w", err), nil
	} else {

		return n, nil, n.clients.Close
	}
}

func (n *Node) initLedger() (*ledger.Ledger, error) {
	// TODO init events here
	// events.Init() //Init event hub

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
