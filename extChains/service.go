package extChains

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
)

type ExtChains interface {
}

type service struct {
	clients []eth.Client
}

type ChainsConfig []*eth.Config

func NewService(ctx context.Context, chainsConfig ChainsConfig) (*service, error) {
	s := &service{
		clients: make([]eth.Client, 0, len(chainsConfig)),
	}
	for _, cfg := range chainsConfig {
		if client, err := eth.NewClient(ctx, cfg); err != nil {

			return nil, err
		} else {
			s.clients = append(s.clients, client)
		}

	}
}
