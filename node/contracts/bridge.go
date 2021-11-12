package contracts

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
)

var (
	ErrBridgeNotSupported = errors.New("bridge not supported")
)

type Bridge interface {
	GetBridge(chainId *big.Int) (*wrappers.Bridge, error)
	GetBridgeAddress(chainId *big.Int) (common.Address, error)
}

func NewBridge(clients *extChains.Clients) (Bridge, error) {
	chainsLen := len(config.Bridge.Chains)
	b := &bridge{
		addressMap: make(map[uint64]common.Address, chainsLen),
		bridgeMap:  make(map[uint64]*wrappers.Bridge, chainsLen),
	}

	for _, chain := range config.Bridge.Chains {
		if client, err := clients.GetEthClient(new(big.Int).SetUint64(chain.Id)); err != nil {
			err = fmt.Errorf("chain [%d] bridge contract get client error: %w", chain.Id, err)
			logrus.Error(err)

			return nil, err

		} else if chainBridge, err := wrappers.NewBridge(chain.BridgeAddress, client); err != nil {
			err = fmt.Errorf("chain [%d] bridge contract not initialized on error: %w", chain.Id, err)
			logrus.Error(err)

			return nil, err
		} else {

			b.addressMap[chain.Id] = chain.ForwarderAddress
			b.bridgeMap[chain.Id] = chainBridge
		}

	}
	return b, nil
}

type bridge struct {
	addressMap map[uint64]common.Address
	bridgeMap  map[uint64]*wrappers.Bridge
}

func (b bridge) GetBridge(chainId *big.Int) (*wrappers.Bridge, error) {

	if chainBridge, ok := b.bridgeMap[chainId.Uint64()]; !ok {

		return nil, fmt.Errorf("%w for chain [%s]", ErrBridgeNotSupported, chainId.String())
	} else {
		return chainBridge, nil
	}

}

func (b bridge) GetBridgeAddress(chainId *big.Int) (common.Address, error) {
	if address, ok := b.addressMap[chainId.Uint64()]; !ok {

		return common.Address{}, fmt.Errorf("%w for chain [%s]", ErrBridgeNotSupported, chainId.String())
	} else {
		return address, nil
	}

}
