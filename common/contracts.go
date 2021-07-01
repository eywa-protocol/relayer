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
	logrus.Printf("Eth ADDRESS for pk provided %x", fromAddress)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logrus.Error(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logrus.Error(err)
	}
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		logrus.Error(err)
	}
	auth, err = bind.NewKeyedTransactorWithChainID(privateKey, chainId)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	//NOTE!!! TODO: should be adjusted for optimal gas cunsumption
	auth.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))
	return
}
