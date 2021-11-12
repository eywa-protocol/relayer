package extChains

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
	"math/big"
)

type Clients struct {
	ethClients map[uint64]eth.Client
}

type ClientConfigs []*eth.Config

func NewClients(ctx context.Context, clientConfigs ClientConfigs) (*Clients, error) {
	s := &Clients{
		ethClients: make(map[uint64]eth.Client, len(clientConfigs)),
	}
	for _, cfg := range clientConfigs {
		if _, exists := s.ethClients[cfg.Id]; exists {

			continue
		} else if client, err := eth.NewClient(ctx, cfg); err != nil {

			return nil, err
		} else {

			s.ethClients[cfg.Id] = client
		}
	}
	return s, nil
}

func (s Clients) IsEth(chainId *big.Int) bool {
	_, ok := s.ethClients[chainId.Uint64()]
	return ok
}

func (s *Clients) GetEthClient(chainId *big.Int) (eth.Client, error) {
	if s.IsEth(chainId) {

		return s.ethClients[chainId.Uint64()], nil
	} else {
		err := fmt.Errorf("chain id [%s] %w", chainId.String(), ErrUnsupported)
		logrus.Error(err)

		return nil, err
	}
}

func (s Clients) AddWatcher(chainId *big.Int, watcher eth.ContractWatcher) {
	if client, exists := s.ethClients[chainId.Uint64()]; exists {

		client.AddWatcher(watcher)
	} else {

		logrus.Warnf("skip add watcher for unsuported chain [%s]", chainId.String())
	}
}

func (s *Clients) Close() {
	for id, client := range s.ethClients {
		client.Close()
		delete(s.ethClients, id)
	}
}
