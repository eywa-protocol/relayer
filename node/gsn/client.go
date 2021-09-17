package gsn

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Client struct {
	EthClient  *ethclient.Client
	ChainCfg   *config.GsnChain
	owner      *bind.TransactOpts
	forwarder  *wrappers.Forwarder
	currentUrl string
}

func NewClient(chain *config.GsnChain, skipUrl string) (client Client, err error) {

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

	forwarder, err := wrappers.NewForwarder(chain.ForwarderAddress, c)
	if err != nil {
		err = fmt.Errorf("init forwarder [%s] error: %w", chain.ForwarderAddress, err)
		return
	}

	owner, err := bind.NewKeyedTransactorWithChainID(chain.EcdsaKey, chain.ChainId)
	if err != nil {
		err = fmt.Errorf("init owner [%s] error: %w", chain.ForwarderAddress, err)
		return
	}

	return Client{
		EthClient:  c,
		ChainCfg:   chain,
		owner:      owner,
		forwarder:  forwarder,
		currentUrl: url,
	}, nil
}
