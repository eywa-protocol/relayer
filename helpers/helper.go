package helpers

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type OracleRequest struct {
	RequestType    string
	Bridge         common.Address
	RequestId      [32]byte
	Selector       []byte
	ReceiveSide    common.Address
	OppositeBridge common.Address
}

type TransactSession struct {
	TransactOpts *bind.TransactOpts // Ethereum account to send the transaction from
	CallOpts     *bind.CallOpts     // Nonce to use for the transaction execution (nil = use pending state)
	*sync.Mutex
	client ethclient.Client
}

func FilterOracleRequestEvent(client ethclient.Client, start uint64, contractAddress common.Address) (oracleRequest OracleRequest, err error) {
	mockFilterer, err := wrappers.NewBridge(contractAddress, &client)
	if err != nil {
		return
	}

	it, err := mockFilterer.FilterOracleRequest(&bind.FilterOpts{Start: start, Context: context.Background()})
	if err != nil {
		return
	}

	for it.Next() {
		logrus.Trace("OracleRequest Event", it.Event.Raw)
		if it.Event != nil {
			oracleRequest = OracleRequest{
				RequestType:    it.Event.RequestType,
				Bridge:         it.Event.Bridge,
				RequestId:      it.Event.RequestId,
				Selector:       it.Event.Selector,
				ReceiveSide:    it.Event.ReceiveSide,
				OppositeBridge: it.Event.OppositeBridge,
			}
		}
	}
	return
}

func WaitTransaction(client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error
	for {
		receipt, err = client.TransactionReceipt(context.Background(), tx.Hash())
		if receipt == nil || err == ethereum.NotFound {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("transaction %s failed: %v", tx.Hash().Hex(), err)
		}
		break
	}
	if receipt.Status != 1 {
		return nil, fmt.Errorf("failed transaction: %s", tx.Hash().Hex())
	}
	return receipt, nil
}

func WaitTransactionDeadline(client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():

			return nil, fmt.Errorf("wait for transaction %s timed out", txHash.Hex())
		default:

			if receipt, err = client.TransactionReceipt(context.Background(), txHash); receipt == nil ||
				errors.Is(err, ethereum.NotFound) {

				time.Sleep(time.Millisecond * 500)
				continue
			} else if err != nil {

				return nil, fmt.Errorf("transaction %s failed: %v", txHash.Hex(), err)
			} else if receipt.Status != 1 {

				return nil, fmt.Errorf("failed transaction: %s", txHash.Hex())
			} else {

				return receipt, nil
			}
		}
	}
}

func WaitTransactionWithRetry(client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error
	for {
		receipt, err = client.TransactionReceipt(context.Background(), tx.Hash())
		if receipt == nil || err == ethereum.NotFound {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("transaction %s failed: %v", tx.Hash().Hex(), err)
		}
		break
	}
	// TODO: confirm that it's always possible to get code message with empty from address and nil block number
	if receipt.Status != 1 {
		msg := ethereum.CallMsg{
			To:   tx.To(),
			Data: tx.Data(),
		}
		code, err := client.CallContract(context.Background(), msg, nil)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("transaction %s failed: %s", tx.Hash().Hex(), code)
	}
	return receipt, nil
}

var GasUsed = big.NewInt(0)

func WaitForBlockCompletation(client *ethclient.Client, hash string) (int, *types.Receipt) {
	ctx, chancel := context.WithTimeout(context.Background(), time.Second*60)
	defer chancel()
	transaction := make(chan *types.Receipt)
	go func(context context.Context, client *ethclient.Client) {
		for {
			statusCode := -1
			txHash := common.HexToHash(hash)
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
	}(ctx, client)
	select {
	case tx := <-transaction:
		if tx != nil {

			txd, _, _ := client.TransactionByHash(context.Background(), tx.TxHash)
			total := new(big.Int).Mul(txd.GasPrice(), new(big.Int).SetUint64(tx.GasUsed))
			GasUsed.Add(GasUsed, total)
			// GasUsed += tx.CumulativeGasUsed
			return int(tx.Status), tx
		}
		return -1, nil
	}
}
