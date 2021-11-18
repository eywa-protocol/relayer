package extChains

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
)

type Clients struct {
	mx         *sync.Mutex
	ethClients map[uint64]eth.Client
}

type ClientConfigs []*eth.Config

func NewClients(ctx context.Context, clientConfigs ClientConfigs) (*Clients, error) {
	s := &Clients{
		mx:         new(sync.Mutex),
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

func (s *Clients) IsEth(chainId *big.Int) bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	_, ok := s.ethClients[chainId.Uint64()]
	return ok
}

func (s *Clients) GetEthClient(chainId *big.Int) (eth.Client, error) {
	s.mx.Lock()
	if client, ok := s.ethClients[chainId.Uint64()]; ok {
		s.mx.Unlock()
		return client, nil
	} else {
		s.mx.Unlock()
		err := fmt.Errorf("chain id [%s] %w", chainId.String(), ErrUnsupported)
		logrus.Error(err)
		return nil, err
	}
}

func (s Clients) AddWatcher(chainId *big.Int, watcher eth.ContractWatcher) {
	s.mx.Lock()
	if client, exists := s.ethClients[chainId.Uint64()]; exists {
		s.mx.Unlock()
		client.AddWatcher(watcher)
	} else {
		s.mx.Unlock()
		logrus.Warnf("skip add watcher for unsuported chain [%s]", chainId.String())
	}
}

func (s *Clients) Close() {
	s.mx.Lock()
	defer s.mx.Unlock()
	for id, client := range s.ethClients {
		client.Close()
		delete(s.ethClients, id)
	}
}

func (s *Clients) All() *map[uint64]eth.Client {
	return &s.ethClients
}
