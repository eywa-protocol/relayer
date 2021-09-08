package config

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func dial(url string) (*ethclient.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if client, err := ethclient.DialContext(ctx, url); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}
