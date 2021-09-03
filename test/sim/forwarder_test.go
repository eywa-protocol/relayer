package sim

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

func TestForwarderVerify(t *testing.T) {

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdRequestTypedData)
	require.NoError(t, err)
	domainSeparator, err := fwdRequestTypedData.HashStruct("EIP712Domain", fwdRequestTypedData.Domain.Map())
	require.NoError(t, err)
	dsep, _ := common2.FromHex(domainSeparator.String())
	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", "address from,address to,uint256 value,uint256 gas,uint256 nonce,bytes data")
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)
	err = forwarder.Verify(&bind.CallOpts{}, *forwarederRequest, dsep, reqTypeHashBytes32, nil, typedDataSignature)
	require.NoError(t, err)
}

func TestMyMethods(t *testing.T) {
	for _, mt := range nodeRegistryABI.Methods {
		t.Log(mt.Sig, mt.Inputs)
	}
}

func TestForwardExecute(t *testing.T) {

	blsPubKey = string(GenRandomBytes(256))
	randWallet = common.BytesToAddress(GenRandomBytes(20))
	pool = common.BytesToAddress(GenRandomBytes(20))
	bridgeOwner = common.BytesToAddress(GenRandomBytes(20))

	createNodeRegistryData = wrappers.NodeRegistryNode{
		Owner:         bridgeOwner,
		Pool:          pool,
		NodeIdAddress: randWallet,
		BlsPubKey:     blsPubKey,
		NodeId:        big.NewInt(0),
	}

	testForwardCreateNodeRegistryABIPacked, err = nodeRegistryABI.Pack("addNode", &createNodeRegistryData)
	if err != nil {
		panic(err)
	}

	//testForwardCreateNodeRegistryABIPacked = []byte("dcdss")

	logrus.Print("testForwardCreateNodeRegistryABIPacked LENGTH ", len(testForwardCreateNodeRegistryABIPacked))

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
		Data:  testForwardCreateNodeRegistryABIPacked,
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
			"data":  testForwardCreateNodeRegistryABIPacked,
		},
	}

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdRequestTypedData)
	require.NoError(t, err)

	domainSeparator, err := fwdRequestTypedData.HashStruct("EIP712Domain", fwdRequestTypedData.Domain.Map())
	require.NoError(t, err)

	dsep, _ := common2.FromHex(domainSeparator.String())
	reqType, err := forwarder.GENERICPARAMS(&bind.CallOpts{})
	require.NoError(t, err)

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", reqType)
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	//_, err := testForward.Foo(owner, big.NewInt(42))
	//require.NoError(t, err)

	resW, err := forwarder.Execute(owner, *forwarederRequest, dsep, reqTypeHashBytes32, nil, typedDataSignature)
	backend.Commit()
	require.NoError(t, err)

	t.Log(testForward.Val(&bind.CallOpts{}))
	t.Log(testForward.Sender(&bind.CallOpts{}))
	sAddr, _ := testForward.Sender(&bind.CallOpts{})
	require.Equal(t, signerAddress, sAddr)
	strRes, _ := testForward.Str(&bind.CallOpts{})
	require.Equal(t, strRes, blsPubKey)
	t.Log("GAS USED", resW.Gas())

	// node, err := testForward.Val(&bind.CallOpts{}, createNodeData.nodeIdAddress)
	// require.NotNil(t, node)
	// t.Log(createNodeData.nodeIdAddress)
	// t.Log(node.NodeIdAddress)
	// require.Equal(t, createNodeData.nodeIdAddress, node.NodeIdAddress)
	// nodesCountAfterTest := getNodesCount()

	// t.Log(nodesCountBeforeTest)
	// t.Log(nodesCountAfterTest)

	// require.True(t, (nodesCountAfterTest-nodesCountBeforeTest == 1))
}

func TestForwardExecute2(t *testing.T) {
	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdRequestTypedData)
	require.NoError(t, err)

	domainSeparator, err := fwdRequestTypedData.HashStruct("EIP712Domain", fwdRequestTypedData.Domain.Map())
	require.NoError(t, err)

	dsep, _ := common2.FromHex(domainSeparator.String())
	reqType, err := forwarder.GENERICPARAMS(&bind.CallOpts{})
	require.NoError(t, err)

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", reqType)
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	//_, err := testForward.Foo(owner, big.NewInt(42))
	//require.NoError(t, err)
	typedDataSignature[31] = 2
	resW, err := forwarder.Execute(owner, *forwarederRequest, dsep, reqTypeHashBytes32, nil, typedDataSignature)
	backend.Commit()
	require.NoError(t, err)

	t.Log(testForward.Val(&bind.CallOpts{}))
	t.Log(testForward.Sender(&bind.CallOpts{}))
	sAddr, _ := testForward.Sender(&bind.CallOpts{})
	require.Equal(t, signerAddress, sAddr)
	strRes, _ := testForward.Str(&bind.CallOpts{})
	require.Equal(t, strRes, blsPubKey)
	t.Log("GAS USED", resW.Gas())

	// node, err := testForward.Val(&bind.CallOpts{}, createNodeData.nodeIdAddress)
	// require.NotNil(t, node)
	// t.Log(createNodeData.nodeIdAddress)
	// t.Log(node.NodeIdAddress)
	// require.Equal(t, createNodeData.nodeIdAddress, node.NodeIdAddress)
	// nodesCountAfterTest := getNodesCount()

	// t.Log(nodesCountBeforeTest)
	// t.Log(nodesCountAfterTest)

	// require.True(t, (nodesCountAfterTest-nodesCountBeforeTest == 1))
}
