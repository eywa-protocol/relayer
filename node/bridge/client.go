package bridge

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Client struct {
	EthClient            *ethclient.Client
	ChainCfg             *config.BridgeChain
	EcdsaKey             *ecdsa.PrivateKey
	Bridge               wrappers.BridgeSession
	BridgeFilterer       wrappers.BridgeFilterer
	NodeRegistry         wrappers.NodeRegistrySession
	NodeRegistryFilterer wrappers.NodeRegistryFilterer
	Forwarder        wrappers.Forwarder

	currentUrl string
}

func NewClient(chain *config.BridgeChain, skipUrl string) (client Client, err error) {

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
	nodeRegistry, err := wrappers.NewNodeRegistry(chain.NodeRegistryAddress, c)
	if err != nil {
		err = fmt.Errorf("init node registry [%s] error: %w", chain.NodeRegistryAddress, err)
		return
	}

	nodeListFilterer, err := wrappers.NewNodeRegistryFilterer(chain.NodeRegistryAddress, c)
	if err != nil {
		err = fmt.Errorf("init nodelist filter [%s] error: %w", chain.NodeRegistryAddress, err)
		return
	}

	forwarder, err := wrappers.NewForwarder(chain.ForwarderAddress, c)
	if err != nil {
		err = fmt.Errorf("init forwarder caller [%s] error: %w", chain.BridgeAddress, err)
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
		NodeRegistry: wrappers.NodeRegistrySession{
			Contract:     nodeRegistry,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
		},
		BridgeFilterer:       *bridgeFilterer,
		NodeRegistryFilterer: *nodeListFilterer,
		Forwarder:        *forwarder,
		currentUrl:           url,
	}, nil
}
