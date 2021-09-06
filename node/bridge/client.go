package bridge

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var (
	ErrGetEthClient = errors.New("get eth client error")
)

type Client struct {
	EthClient        *ethclient.Client
	ChainCfg         *config.BridgeChain
	EcdsaKey         *ecdsa.PrivateKey
	Bridge           wrappers.BridgeSession
	BridgeFilterer   wrappers.BridgeFilterer
	NodeList         wrappers.NodeListSession
	NodeListFilterer wrappers.NodeListFilterer

	currentUrl string
}

func NewClient(chain *config.BridgeChain, skipUrl string) (client Client, err error) {

	client = Client{
		ChainCfg: chain,
		EcdsaKey: chain.EcdsaKey,
	}
	c, url, err := chain.GetEthClient(skipUrl)
	if err != nil {
		err = fmt.Errorf("chain[%d] %s: %w", chain.Id, ErrGetEthClient.Error(), err)
		logrus.Error(err)
		return client, ErrGetEthClient
	}
	chain.ChainId, err = c.ChainID(context.Background())
	if err != nil {
		err = fmt.Errorf("chain[%d] get chain id from blockchain error: %w", chain.Id, err)
		logrus.Error(err)
		return client, ErrGetEthClient
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
