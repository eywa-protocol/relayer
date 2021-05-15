package common

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"math/big"
)

import (
	"context"
)

/*
type TransactSession struct {
	TransactOpts *bind.TransactOpts // Ethereum account to send the transaction from
	CallOpts     *bind.CallOpts     // Nonce to use for the transaction execution (nil = use pending state)
	*sync.Mutex
	Client  ethclient.Client
	Backend bind.ContractBackend
}

func GetContractByAddressNetwork1(config config.AppConfig, ethClient *ethclient.Client) (out *wrappers.Bridge, err error) {
	return wrappers.NewBridge(common.HexToAddress(config.BRIDGE_ADDRESS_NETWORK1), ethClient)
}
*/
func CustomAuth(client *ethclient.Client, privateKey *ecdsa.PrivateKey) (auth *bind.TransactOpts) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logrus.Error("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logrus.Error(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Error(err)
	}
	auth = bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasPrice = gasPrice
	return
}

/*
func CustomAuth2(client *ethclient.Client, privateKey *ecdsa.PrivateKey) (auth *bind.TransactOpts) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logrus.Error("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logrus.Error(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Error(err)
	}
	auth = bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce - 1))
	auth.Value = big.NewInt(0) // in wei
	auth.GasPrice = gasPrice
	return
}

func NewTransactSession(backend *ethclient.Client, privateKey *ecdsa.PrivateKey) (*TransactSession, error) {
	transactOpts := CustomAuth(backend, privateKey)
	transactOpts.Context = context.Background()
	nonce, err := backend.PendingNonceAt(context.Background(), transactOpts.From)
	if err != nil {
		return nil, fmt.Errorf("pending nonce at %s error: %v", transactOpts.From.Hex(), err)
	}
	transactOpts.Nonce = (&big.Int{}).SetUint64(nonce)
	txSession := TransactSession{
		TransactOpts: transactOpts,
		CallOpts:     &bind.CallOpts{},
		Mutex:        &sync.Mutex{},
		Client:       *backend,
	}
	return &txSession, nil
}

func (txSession *TransactSession) WaitTransactionOLD(tx *types.Transaction) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error
	for {
		receipt, err = txSession.Client.TransactionReceipt(context.Background(), tx.Hash())
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

func (txSession *TransactSession) WaitTransaction(tx *types.Transaction) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error
	for {
		receipt, err = txSession.Client.TransactionReceipt(context.Background(), tx.Hash())
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
		code, err := txSession.Backend.CallContract(context.Background(), msg, nil)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("transaction %s failed: %s", tx.Hash().Hex(), code)
	}
	return receipt, nil
}

func WaitTransactionEXTRA(ctx context.Context, c ethclient.Client, txHash common.Hash) (*types.Receipt, error) {

	logrus.Info("WaitTransaction ", "txHash", txHash)

	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	//logger := logrus.New()

	for {

		logrus.Info("WaitTransaction ", "txHash", txHash)

		receipt, err := c.TransactionReceipt(ctx, txHash)
		if receipt != nil {
			return receipt, nil
		}
		if err != nil {
			logrus.Error("Receipt retrieval failed", "err", err)
		} else {
			logrus.Trace("Transaction not yet mined")
		}
		// Wait for the next round.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}
*/
