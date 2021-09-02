package sim

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var backend *backends.SimulatedBackend

var owner *bind.TransactOpts

var forwarder *wrappers.Forwarder
var testTarget *wrappers.TestTarget

var bridge *wrappers.Bridge
var testForward *wrappers.TestForward
var mockDexPool *wrappers.MockDexPool
var testForwardAddNodeABIPacked []byte

var testForwardABI = common2.MustGetABI(wrappers.TestForwardABI)
var testTargetABI = common2.MustGetABI(wrappers.TestTargetABI)

var fwdRequestTypedData signer.TypedData

var testTargetAddress,
	testForwarderTargetAddress,
	testfForwarderAddress,
	ownerAddress,
	miniForwarderAddress,
	forwarderAddress,
	signerAddress,
	testForwardAddress,
	forwarderTestAddress,
	bridgeAddress,
	mockDexPooolAddress common.Address
var forwarederRequest *wrappers.IForwarderForwardRequest
var blsPubKey string
var err error
var ownerKey, signerKey *ecdsa.PrivateKey

var ctx context.Context
var Domain map[string]json.RawMessage
var domainChainIDAsString map[string]json.RawMessage
var Msg map[string]json.RawMessage

var createNodeData *createNodeDataTypw

type createNodeDataTypw struct {
	nodeWallet    common.Address
	nodeIdAddress common.Address
	blsPubKey     string
}

func init() {
	ctx = context.Background()

	ownerKey, _ = crypto.GenerateKey()

	signerKey, _ = crypto.GenerateKey()

	signerAddress = crypto.PubkeyToAddress(signerKey.PublicKey)

	ownerAddress = crypto.PubkeyToAddress(ownerKey.PublicKey)

	genesis := core.GenesisAlloc{
		ownerAddress: {Balance: new(big.Int).SetInt64(math.MaxInt64)},
	}
	backend = backends.NewSimulatedBackend(genesis, math.MaxInt64)

	owner, err = bind.NewKeyedTransactorWithChainID(ownerKey, big.NewInt(1337))
	if err != nil {
		panic(err)
	}

	forwarderAddress, _, forwarder, err = wrappers.DeployForwarder(owner, backend)
	if err != nil {
		panic(err)
	}
	backend.Commit()

	testTargetAddress, _, testTarget, err = wrappers.DeployTestTarget(owner, backend)
	if err != nil {
		panic(err)
	}
	backend.Commit()

	testForwardAddress, _, testForward, err = wrappers.DeployTestForward(owner, backend, forwarderAddress)
	if err != nil {
		panic(err)
	}
	backend.Commit()

	bridgeAddress, _, bridge, err = wrappers.DeployBridge(owner, backend, testForwardAddress)
	if err != nil {
		panic(err)
	}
	backend.Commit()

	// mockDexPooolAddress, _, mockDexPool, err = wrappers.DeployMockDexPool(owner, backend, testForwardAddress)
	_, _, mockDexPool, err = wrappers.DeployMockDexPool(owner, backend, testForwardAddress)
	if err != nil {
		panic(err)
	}
	backend.Commit()

	// _, err := bridge.UpdateDexBind(owner, mockDexPooolAddress, true)
	// if err != nil {
	// 	panic(err)
	// }
	// backend.Commit()

	logrus.Info("testForwardAddress: ", testForwardAddress)
	initForwarderContractCall()
}

func initForwarderContractCall() {

	blsPubKey = string(GenRandomBytes(256))
	randWallet := common.BytesToAddress(GenRandomBytes(3))

	createNodeData = &createNodeDataTypw{
		nodeWallet:    randWallet,
		nodeIdAddress: randWallet,
		blsPubKey:     blsPubKey,
	}

	logrus.Print("createNodeData.blsPubKey ", createNodeData.blsPubKey)
	logrus.Print("createNodeData.nodeIdAddress ", createNodeData.nodeIdAddress)
	logrus.Print("createNodeData.nodeWallet ", createNodeData.nodeWallet)


	testForwardAddNodeABIPacked, err = testForwardABI.Pack("foo", big.NewInt(42), blsPubKey)
	if err != nil {
		panic(err)
	}

	logrus.Print("LENGTH ", len(testForwardAddNodeABIPacked))

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {
		panic(err)
	}

	forwarederRequest = &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    testForwardAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  testForwardAddNodeABIPacked,
	}

	fwdRequestTypedData = signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{{
				Name: "verifyingContract", Type: "address"}},
			"ForwardRequest": []signer.Type{
				{
					Name: "from", Type: "address"},
				{
					Name: "to", Type: "address"},
				{
					Name: "value", Type: "uint256"},
				{
					Name: "gas", Type: "uint256"},
				{
					Name: "nonce", Type: "uint256"},
				{
					Name: "data", Type: "bytes"},
			},
		},
		Domain: signer.TypedDataDomain{
			VerifyingContract: forwarderAddress.String(),
		},
		PrimaryType: "ForwardRequest",
		Message: signer.TypedDataMessage{
			"from":  forwarederRequest.From.String(),
			"to":    forwarederRequest.To.String(),
			"value": forwarederRequest.Value.String(),
			"gas":   forwarederRequest.Gas.String(),
			"nonce": forwarederRequest.Nonce.String(),
			"data":  testForwardAddNodeABIPacked,
		},
	}

}

func GenRandomBytes(size int) (blk []byte) {
	rand.Seed(time.Now().UnixNano())
	blk = make([]byte, size)
	_, _ = rand.Read(blk)
	return
}

func getSignerNonceFromForwarder() (nonce *big.Int) {
	nonce, err = forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {
		panic(err)
	}
	return
}
