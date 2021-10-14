package eth

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

func (c *client) ChainID(ctx context.Context) (*big.Int, error) {
	if _, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return c.chainId, nil
	}
}

func (c *client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.BalanceAt(callCtx, account, blockNumber)
	}
}

func (c *client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.TransactionReceipt(callCtx, txHash)
	}
}

func (c *client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.CodeAt(callCtx, contract, blockNumber)
	}
}

func (c client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.CallContract(callCtx, call, blockNumber)
	}
}

func (c client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.PendingCodeAt(callCtx, account)
	}
}

func (c client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return 0, err
	} else {
		return client.PendingNonceAt(callCtx, account)
	}
}

func (c client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.SuggestGasPrice(callCtx)
	}
}

func (c client) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return 0, err
	} else {
		return client.EstimateGas(callCtx, call)
	}
}

func (c client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return err
	} else {
		return client.SendTransaction(callCtx, tx)
	}

}

func (c *client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.FilterLogs(callCtx, query)
	}
}

func (c *client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, err
	} else {
		return client.SubscribeFilterLogs(callCtx, query, ch)
	}
}
