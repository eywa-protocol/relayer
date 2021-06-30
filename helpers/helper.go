package helpers

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"sync"
	"time"

	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
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

/**
* Catch event from first side
 */
// UNUSED CODE
/*func ListenOracleRequest(
	clientNetwork_1 *ethclient.Client,
	clientNetwork_2 *ethclient.Client,
	proxyNetwork_1 common.Address,
	proxyNetwork_2 common.Address) (oracleRequest OracleRequest, tx *types.Transaction, err error) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork_1, clientNetwork_1)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeOracleRequest)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchOracleRequest(opt, channel)
	defer sub.Unsubscribe()
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case _ = <-sub.Err():
				break
			case event := <-channel:
				logrus.Tracef("OracleRequest id: %v type: %v\n", event.RequestId, event.RequestType)

				privateKey, err := crypto.HexToECDSA("95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc")
				if err != nil {
					logrus.Fatal(err)
				}
				publicKey := privateKey.Public()
				publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
				if !ok {
					logrus.Fatal("error casting public key to ECDSA")
				}
				fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
				nonce, err := clientNetwork_2.PendingNonceAt(context.Background(), fromAddress)
				if err != nil {
					logrus.Fatal(err)
				}
				gasPrice, err := clientNetwork_2.SuggestGasPrice(context.Background())
				if err != nil {
					logrus.Fatal(err)
				}
				auth := bind.NewKeyedTransactor(privateKey)
				auth.Nonce = big.NewInt(int64(nonce))
				auth.Value = big.NewInt(0)     // in wei
				auth.GasLimit = uint64(300000) // in units
				auth.GasPrice = gasPrice

				instance, err := wrappers.NewBridge(proxyNetwork_2, clientNetwork_2)
				if err != nil {
					logrus.Fatal(err)
				}

				/**   TODO: apporove that tx was real  /
				/**   TODO: apporove that tx was real  /
				/**   TODO: apporove that tx was real  /
				/**   TODO: apporove that tx was real  /

				oracleRequest = OracleRequest{
					RequestType:    event.RequestType,
					Bridge:         event.Bridge,
					RequestId:      event.RequestId,
					Selector:       event.Selector,
					ReceiveSide:    event.ReceiveSide,
					OppositeBridge: event.OppositeBridge,
				}
				/** Invoke bridge on another side /
				tx, err = instance.ReceiveRequestV2(auth, "", nil, oracleRequest.Selector, [32]byte{}, oracleRequest.ReceiveSide)
				if err != nil {
					logrus.Fatal(err)
				}

				logrus.Tracef("tx in first chain has been triggered :  %x", tx.Hash())

			}
		}
	}()
	return
}*/

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
	//TODO: confirm that it's always possible to get code message with empty from address and nil block number
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
			//tx.BlockNumber.String()
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
			//GasUsed += tx.CumulativeGasUsed
			return int(tx.Status), tx
		}
		return -1, nil
	}
}
