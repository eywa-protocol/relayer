package sim

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"github.com/stretchr/testify/require"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
	"testing"
	"time"
)

func Test_methods(t *testing.T) {

	//for _, mt := range nodeListABI.Methods {
	//	t.Log(mt.Sig, mt.Inputs)
	//}
	blsPubKey1 := string(GenRandomBytes(256))
	randWallet1 := common.BytesToAddress(GenRandomBytes(20))
	nodePoll := common.BytesToAddress(GenRandomBytes(20))
	nodeOwner := common.BytesToAddress(GenRandomBytes(20))

	t.Log(blsPubKey1)
	t.Log(randWallet1)
	t.Log(nodePoll)
	t.Log(nodeOwner)
	nodeData := wrappers.NodeRegistryNode{
		Owner:         nodeOwner,
		Pool:          nodePoll,
		NodeIdAddress: randWallet1,
		BlsPubKey:     blsPubKey1,
		NodeId:        big.NewInt(0),
	}

	//createNodeABIPacked, err := nodeRegistryABI.Pack("addNode",
	//	nodeData)
	//	require.NoError(t, err)

	MintTokens(t)
	deadLine := big.NewInt(time.Now().Unix() + 100)

	v, r, s := common2.SignErc20Permit(signerKey, "EYWA", "1", big.NewInt(1337),
		eywaTokenAddress, signerAddress, nodeRegistryAddress, big.NewInt(1e18), big.NewInt(0), deadLine)

	createNodeABIPacked, err := nodeRegistryABI.Pack("createRelayer", nodeData, deadLine, v, r, s)
	require.NoError(t, err)

	//createNodeABIPacked, err := nodeListABI.Pack("createNode1",
	//	createNodeRegistryData.NodeWallet,
	//	createNodeRegistryData.NodeIdAddress,
	//	createNodeRegistryData.BlsPubKey)
	//require.NoError(t, err)

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {
		panic(err)
	}
	createNodeForwarederRequest := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    nodeRegistryAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  createNodeABIPacked,
	}

	fwdRequestTypedDataNew := signer.TypedData{
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
			"from":  createNodeForwarederRequest.From.String(),
			"to":    createNodeForwarederRequest.To.String(),
			"value": createNodeForwarederRequest.Value.String(),
			"gas":   createNodeForwarederRequest.Gas.String(),
			"nonce": createNodeForwarederRequest.Nonce.String(),
			"data":  createNodeABIPacked,
		},
	}

	nodesCountBeforeTest := getNodesCount()
	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdRequestTypedDataNew)
	require.NoError(t, err)

	domainSeparator, err := fwdRequestTypedDataNew.HashStruct("EIP712Domain", fwdRequestTypedDataNew.Domain.Map())
	require.NoError(t, err)

	dsep, _ := common2.FromHex(domainSeparator.String())
	reqType, err := forwarder.GENERICPARAMS(&bind.CallOpts{})
	require.NoError(t, err)

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", reqType)
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	_, err = forwarder.Execute(owner, *createNodeForwarederRequest, dsep, reqTypeHashBytes32, nil, typedDataSignature)
	backend.Commit()
	require.NoError(t, err)

	node, err := nodeRegistry.GetNode(&bind.CallOpts{}, createNodeRegistryData.NodeIdAddress)
	require.NotNil(t, node)
	t.Log(nodeData.NodeIdAddress)
	t.Log(nodeData.NodeIdAddress)
	require.Equal(t, createNodeRegistryData.NodeIdAddress, node.NodeIdAddress)
	nodesCountAfterTest := getNodesCount()

	t.Log(nodesCountBeforeTest)
	t.Log(nodesCountAfterTest)

	require.True(t, (nodesCountAfterTest-nodesCountBeforeTest == 1))

}

func MintTokens(t *testing.T) {
	_, err := eywaToken.Mint(owner, signerAddress, big.NewInt(1e18))
	if err != nil {
		err = fmt.Errorf("Mint: %w", err)
		return
	}

}
