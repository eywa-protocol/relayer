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
	ErrGetEthClient = errors.New("eth client not initialized")
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

func (c *Client) RecreateContractsAndFilters() (err error) {

	if c.EthClient == nil {
		return ErrGetEthClient
	}

	c.ChainCfg.ChainId, err = c.EthClient.ChainID(context.Background())
	if err != nil {
		err = fmt.Errorf("chain[%d] get chain id from blockchain error: %w", c.ChainCfg.Id, err)
		logrus.Error(err)
		return ErrGetEthClient
	}

	bridge, err := wrappers.NewBridge(c.ChainCfg.BridgeAddress, c.EthClient)
	if err != nil {
		err = fmt.Errorf("init bridge [%s] error: %w", c.ChainCfg.BridgeAddress, err)
		return
	}

	bridgeFilterer, err := wrappers.NewBridgeFilterer(c.ChainCfg.BridgeAddress, c.EthClient)
	if err != nil {
		err = fmt.Errorf("init bridge filter [%s] error: %w", c.ChainCfg.BridgeAddress, err)
		return
	}
	nodeList, err := wrappers.NewNodeList(c.ChainCfg.NodeListAddress, c.EthClient)
	if err != nil {
		err = fmt.Errorf("init nodelist [%s] error: %w", c.ChainCfg.BridgeAddress, err)
		return
	}

	nodeListFilterer, err := wrappers.NewNodeListFilterer(c.ChainCfg.NodeListAddress, c.EthClient)
	if err != nil {
		err = fmt.Errorf("init nodelist filter [%s] error: %w", c.ChainCfg.BridgeAddress, err)
		return
	}

	txOpts := common2.CustomAuth(c.EthClient, c.ChainCfg.EcdsaKey)

	c.Bridge = wrappers.BridgeSession{
		Contract:     bridge,
		CallOpts:     bind.CallOpts{},
		TransactOpts: *txOpts,
	}
	c.NodeList = wrappers.NodeListSession{
		Contract:     nodeList,
		CallOpts:     bind.CallOpts{},
		TransactOpts: *txOpts,
	}
	c.BridgeFilterer = *bridgeFilterer
	c.NodeListFilterer = *nodeListFilterer

	return nil
}

func NewClient(chain *config.BridgeChain, skipUrl string) (client Client, err error) {

	client = Client{
		ChainCfg: chain,
		EcdsaKey: chain.EcdsaKey,
	}
	client.EthClient, client.currentUrl, err = chain.GetEthClient(skipUrl)
	if err != nil {
		err = fmt.Errorf("chain[%d] %s: %w", chain.Id, ErrGetEthClient.Error(), err)
		logrus.Error(err)
		return client, ErrGetEthClient
	}
	err = client.RecreateContractsAndFilters()
	return
}
