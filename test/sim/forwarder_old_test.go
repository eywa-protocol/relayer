package sim

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"testing"
)

func TestForwarderVerifyOLD(t *testing.T) {
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

func TestForwarderExecuteOLD(t *testing.T) {
	initForwarderContractCall()
	nodesCountBeforeTest := getNodesCount()
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

	_, err = forwarder.Execute(owner, *forwarederRequest, dsep, reqTypeHashBytes32, nil, typedDataSignature)
	backend.Commit()
	require.NoError(t, err)

	node, err := nodeRegistry.GetNode(&bind.CallOpts{}, createNodeRegistryData.NodeIdAddress)
	require.NotNil(t, node)
	t.Log(createNodeRegistryData.NodeIdAddress)
	t.Log(node.NodeIdAddress)
	require.Equal(t, createNodeRegistryData.NodeIdAddress, node.NodeIdAddress)
	nodesCountAfterTest := getNodesCount()

	t.Log(nodesCountBeforeTest)
	t.Log(nodesCountAfterTest)

	require.True(t, (nodesCountAfterTest-nodesCountBeforeTest == 1))
}
