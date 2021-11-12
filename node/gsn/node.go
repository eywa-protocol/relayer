package gsn

import (
	"context"
	"errors"
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Node struct {
	base.Node
	cMx     *sync.Mutex
	clients *extChains.Clients
	chains  map[uint64]config.GsnChain
}

func RunNode(name, keysPath, listen string, port uint) (err error) {

	nodeKey, err := common2.GetOrGenAndSaveECDSAKey(keysPath, name)
	if err != nil {
		panic(err)
	}

	logrus.Tracef("keyfile %v", keysPath+"/"+name+"-ecdsa.key")

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listen, port))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := libp2p.NewHost(ctx, nodeKey, multiAddr)
	if err != nil {
		panic(err)
	}

	n := &Node{
		Node:   base.Node{Ctx: ctx, Host: h},
		cMx:    new(sync.Mutex),
		chains: make(map[uint64]config.GsnChain, len(config.Gsn.Chains)),
	}

	logrus.Print(len(config.Gsn.Chains), " chains Length")

	clientConfigs := make(extChains.ClientConfigs, 0, len(config.Gsn.Chains))

	for _, chain := range config.Gsn.Chains {
		logrus.Print("CHAIN ", chain, "chain")

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
		n.chains[chain.Id] = *chain
	}

	n.clients, err = extChains.NewClients(ctx, clientConfigs)
	if err != nil {
		err = fmt.Errorf("init eth clients error: %w", err)
		logrus.Error(err)
		return err
	}
	defer n.clients.Close()

	n.Dht, err = n.InitDHT(config.Gsn.BootstrapAddrs)
	if err != nil {
		return fmt.Errorf("init DHT error: %w", err)
	}

	if _, err := gsn.NewServer(n.Ctx, n.Host, n); err != nil {
		return fmt.Errorf("init gsn rpc service error: %w", err)
	}

	runa.Host(h, cancel, &sync.WaitGroup{})

	return
}

func (n *Node) GetOwner(chainId *big.Int) (*bind.TransactOpts, error) {
	if chain, ok := n.chains[chainId.Uint64()]; !ok {

		return nil, fmt.Errorf("chain[%s] error: %w", chainId.String(), errors.New("unsupported now"))
	} else {

		return bind.NewKeyedTransactorWithChainID(chain.EcdsaKey, chainId)
	}
}

func (n *Node) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {
	if chain, ok := n.chains[chainId.Uint64()]; !ok {

		return nil, fmt.Errorf("chain[%s] error: %w", chainId.String(), errors.New("unsupported now"))
	} else if client, err := n.clients.GetEthClient(chainId); err != nil {

		return nil, err
	} else {

		return wrappers.NewForwarder(chain.ForwarderAddress, client)
	}
}
