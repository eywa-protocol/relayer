package eth

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

func (c *client) ChainID(_ context.Context) (*big.Int, error) {
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

func (c *client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	if client, err := c.getClient(); err != nil {
		return nil, false, err
	} else {
		return client.TransactionByHash(callCtx, hash)
	}
}

func (c *client) CallOpt(privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrEcdsaPubCast
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := c.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}
	gasPrice, err := c.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	chainId, err := c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	// NOTE!!! TODO: should be adjusted for optimal gas consumption
	auth.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))

	return auth, nil
}

func (c *client) WaitForBlockCompletion(txHash common.Hash) (int, *types.Receipt) {
	ctx, chancel := context.WithTimeout(context.Background(), time.Second*60)
	defer chancel()
	transaction := make(chan *types.Receipt)
	go func(context context.Context, client *client) {
		for {
			statusCode := -1
			tx, err := client.TransactionReceipt(ctx, txHash)
			// tx.BlockNumber.String()
			if err == nil {
				statusCode = int(tx.Status)
				transaction <- tx
				return
			} else {
				statusCode = -1
			}
			select {
			case <-ctx.Done():
				if statusCode == -1 {
					transaction <- nil
				} else {
					transaction <- tx
				}
				break
			default:
				_ = 1
			}
			time.Sleep(time.Second * 2)
		}
	}(ctx, c)
	select {
	case tx := <-transaction:
		if tx != nil {

			txd, _, _ := c.TransactionByHash(context.Background(), tx.TxHash)
			total := new(big.Int).Mul(txd.GasPrice(), new(big.Int).SetUint64(tx.GasUsed))
			gasUsed := big.NewInt(0)
			gasUsed.Add(gasUsed, total)
			// GasUsed += tx.CumulativeGasUsed
			return int(tx.Status), tx
		}
		return -1, nil
	}
}
