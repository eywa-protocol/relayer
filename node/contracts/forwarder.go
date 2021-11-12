package contracts

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
)

var (
	ErrForwarderNotSupported = errors.New("forwarder not supported")
)

type Forwarder interface {
	GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error)
	GetForwarderAddress(chainId *big.Int) (common.Address, error)
	CanUseGsn(chainId *big.Int) bool
}

func NewForwarder(clients *extChains.Clients) (Forwarder, error) {
	chainsLen := len(config.Bridge.Chains)
	f := &forwarder{
		addressMap:   make(map[uint64]common.Address, chainsLen),
		forwarderMap: make(map[uint64]*wrappers.Forwarder, chainsLen),
	}

	for _, chain := range config.Bridge.Chains {
		if client, err := clients.GetEthClient(new(big.Int).SetUint64(chain.Id)); err != nil {
			err = fmt.Errorf("chain [%d] forwarder contract get client error: %w", chain.Id, err)
			logrus.Error(err)

			return nil, err

		} else if common2.AddressIsZero(chain.ForwarderAddress) || !chain.UseGsn {

			continue
		} else if chainForwarder, err := wrappers.NewForwarder(chain.ForwarderAddress, client); err != nil {
			err = fmt.Errorf("chain [%d] forwarder contract not initialized on error: %w", chain.Id, err)
			logrus.Error(err)

			return nil, err
		} else {

			f.addressMap[chain.Id] = chain.ForwarderAddress
			f.forwarderMap[chain.Id] = chainForwarder
		}

	}
	return f, nil
}

type forwarder struct {
	addressMap   map[uint64]common.Address
	forwarderMap map[uint64]*wrappers.Forwarder
}

func (f *forwarder) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {
	if chainForwarder, ok := f.forwarderMap[chainId.Uint64()]; !ok {

		return nil, fmt.Errorf("%w for chain [%s]", ErrForwarderNotSupported, chainId.String())
	} else {
		return chainForwarder, nil
	}
}

func (f *forwarder) GetForwarderAddress(chainId *big.Int) (common.Address, error) {
	if address, ok := f.addressMap[chainId.Uint64()]; !ok {

		return common.Address{}, fmt.Errorf("%w for chain [%s]", ErrForwarderNotSupported, chainId.String())
	} else {
		return address, nil
	}

}

func (f forwarder) CanUseGsn(chainId *big.Int) bool {
	_, ok := f.forwarderMap[chainId.Uint64()]

	return ok
}
