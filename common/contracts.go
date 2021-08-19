package common

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core"
	"github.com/sirupsen/logrus"
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

func MustGetABI(json string) abi.ABI {
	abi, err := abi.JSON(strings.NewReader(json))
	if err != nil {
		panic("could not parse ABI: " + err.Error())
	}
	return abi
}

func SignTypedData(key ecdsa.PrivateKey, addr common.MixedcaseAddress, typedData core.TypedData) (signature hexutil.Bytes, SignDataRequestHash hexutil.Bytes, err error) {
	//logrus.Print(typedData.Domain.Map())
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, nil, err
	}
	//mainHash := fmt.Sprintf("0x%s", common.Bytes2Hex(typedDataHash))
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256(rawData)
	messages, err := typedData.Format()
	if err != nil {
		return nil, nil, err
	}
	req := &core.SignDataRequest{
		ContentType: core.DataTyped.Mime,
		Rawdata:     rawData,
		Messages:    messages,
		Hash:        sighash,
		Address:     addr}
	signature, err = crypto.Sign(req.Hash, &key)
	if err != nil {
		return nil, nil, err
	}
	signature[64] += 27
	return signature, req.Hash, nil
}

func RawSimTx(client *backends.SimulatedBackend, opts *bind.TransactOpts, contractAddr *common.Address, input []byte) (*types.Transaction, error) {
	var err error

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = client.PendingNonceAt(ensureContext(opts.Context), opts.From)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}
	// Figure out the gas allowance and gas price values
	gasPrice := opts.GasPrice
	if gasPrice == nil {
		gasPrice, err = client.SuggestGasPrice(ensureContext(opts.Context))
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas price: %v", err)
		}
	}
	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		// Gas estimation cannot succeed without code for method invocations
		if contractAddr != nil {
			if code, err := client.PendingCodeAt(ensureContext(opts.Context), *contractAddr); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, errors.New("ErrNoCode")
			}
		}
		// If the contractAddr surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: opts.From, To: contractAddr, Value: value, Data: input}
		gasLimit, err = client.EstimateGas(ensureContext(opts.Context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
	}

	txdata := &types.LegacyTx{
		Nonce: nonce,
		To:    contractAddr,
		Data:  common.CopyBytes(input),
		Gas:   gasLimit,
		// These are initialized below.
		Value:    big.NewInt(0),
		GasPrice: big.NewInt(2000000000),
	}

	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	rawTx = types.NewTx(txdata)
	if opts.Signer == nil {
		return nil, errors.New("no signer to authorize the transaction with")
	}
	signedTx, err := opts.Signer(opts.From, rawTx)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}
