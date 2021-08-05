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
	"math/rand"
	"testing"
)

func Test_SetUintTarget(t *testing.T) {
	setValue := big.NewInt(rand.Int63())
	testTargetSetUintIPacked, err := testTargetABI.Pack("setTestUint",
		setValue)
	require.NoError(t, err)
	//_= testTargetSetUintIPacked

	nonce := getSignerNonceFromForwarder()

	forwarederTestTargetRequest := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    testTargetAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  testTargetSetUintIPacked,
	}

	fwdTestTargetRequestTypedData := signer.TypedData{
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
			"from":  forwarederTestTargetRequest.From.String(),
			"to":    forwarederTestTargetRequest.To.String(),
			"value": forwarederTestTargetRequest.Value.String(),
			"gas":   forwarederTestTargetRequest.Gas.String(),
			"nonce": forwarederTestTargetRequest.Nonce.String(),
			"data":  testTargetSetUintIPacked,
		},
	}

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdTestTargetRequestTypedData)
	require.NoError(t, err)

	testTargetDomainSeparator, err := fwdTestTargetRequestTypedData.HashStruct("EIP712Domain", fwdTestTargetRequestTypedData.Domain.Map())
	require.NoError(t, err)

	testTargetDsep, _ := common2.FromHex(testTargetDomainSeparator.String())

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", "address from,address to,uint256 value,uint256 gas,uint256 nonce,bytes data")

	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	_, err = forwarder.Execute(owner, *forwarederTestTargetRequest, testTargetDsep, reqTypeHashBytes32, nil, typedDataSignature)

	require.NoError(t, err)
	backend.Commit()

	result, err := testTarget.TestUint(&bind.CallOpts{})
	require.Equal(t, setValue, result)
	t.Log(result)
	t.Log(setValue)

}

func Test_SetAddressTarget(t *testing.T) {
	setValue := common.BytesToAddress([]byte("123456789"))
	testTargetSetUintIPacked, err := testTargetABI.Pack("setTestAddress",
		setValue)
	require.NoError(t, err)
	//_= testTargetSetUintIPacked

	nonce := getSignerNonceFromForwarder()

	forwarederTestTargetRequest := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    testTargetAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  testTargetSetUintIPacked,
	}

	fwdTestTargetRequestTypedData := signer.TypedData{
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
			"from":  forwarederTestTargetRequest.From.String(),
			"to":    forwarederTestTargetRequest.To.String(),
			"value": forwarederTestTargetRequest.Value.String(),
			"gas":   forwarederTestTargetRequest.Gas.String(),
			"nonce": forwarederTestTargetRequest.Nonce.String(),
			"data":  testTargetSetUintIPacked,
		},
	}

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdTestTargetRequestTypedData)
	require.NoError(t, err)

	testTargetDomainSeparator, err := fwdTestTargetRequestTypedData.HashStruct("EIP712Domain", fwdTestTargetRequestTypedData.Domain.Map())
	require.NoError(t, err)

	testTargetDsep, _ := common2.FromHex(testTargetDomainSeparator.String())

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", "address from,address to,uint256 value,uint256 gas,uint256 nonce,bytes data")

	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	_, err = forwarder.Execute(owner, *forwarederTestTargetRequest, testTargetDsep, reqTypeHashBytes32, nil, typedDataSignature)

	require.NoError(t, err)
	backend.Commit()

	result, err := testTarget.TesAddress(&bind.CallOpts{})
	require.Equal(t, setValue, result)
	t.Log(result)
	t.Log(setValue)

}

func Test_SetBytesTarget(t *testing.T) {
	//setValue := common.Hex2Bytes("12358459487543985749587439AAAAAAA111111999999999999999999999AA")
	setValue := common.Hex2Bytes("AA12358459487543985749587439AAAAAAA111111999999999999999999999AAaaaaaaaaaaaaaaaaaaaaa")
	//t.Log(testTargetABI.Methods)
	for _, mt := range testTargetABI.Methods {
		t.Log(mt.Sig, mt.Inputs)
	}

	testTargetSetUintIPacked, err := testTargetABI.Pack("setTestBytes",
		setValue)
	require.NoError(t, err)
	_ = testTargetSetUintIPacked

	nonce := getSignerNonceFromForwarder()

	forwarederTestTargetRequest := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    testTargetAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  testTargetSetUintIPacked,
	}

	fwdTestTargetRequestTypedData := signer.TypedData{
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
			"from":  forwarederTestTargetRequest.From.String(),
			"to":    forwarederTestTargetRequest.To.String(),
			"value": forwarederTestTargetRequest.Value.String(),
			"gas":   forwarederTestTargetRequest.Gas.String(),
			"nonce": forwarederTestTargetRequest.Nonce.String(),
			"data":  testTargetSetUintIPacked,
		},
	}

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdTestTargetRequestTypedData)
	require.NoError(t, err)

	testTargetDomainSeparator, err := fwdTestTargetRequestTypedData.HashStruct("EIP712Domain", fwdTestTargetRequestTypedData.Domain.Map())
	require.NoError(t, err)

	testTargetDsep, _ := common2.FromHex(testTargetDomainSeparator.String())

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", "address from,address to,uint256 value,uint256 gas,uint256 nonce,bytes data")

	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	_, err = forwarder.Execute(owner, *forwarederTestTargetRequest, testTargetDsep, reqTypeHashBytes32, nil, typedDataSignature)

	require.NoError(t, err)
	backend.Commit()

	t.Log(setValue)
	result, err := testTarget.BytesData(&bind.CallOpts{})
	require.Equal(t, setValue, result)
	t.Log(result)

}

func Test_SetStringTarget(t *testing.T) {
	setValue := "08d8cb864651e18cc623e2d0fb1f067e943656d175cde3ee4c381d2cb227839c105cdb12d6f3609816790a1246fe5e3c140acbd56c427e8c757b3f0ea7d664a11f9d71dc4bda679e8ae19e5f12ae5b64a4d65adbf8c1425d291def3dbc12ae0e3ee981e1e38aea669f09a85a6b8cfd2724e50919438625c0e9ff4dc291daa535"
	testTargetSetUintIPacked, err := testTargetABI.Pack("setTestString",
		setValue)
	require.NoError(t, err)

	nonce := getSignerNonceFromForwarder()

	forwarederTestTargetRequest := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    testTargetAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  testTargetSetUintIPacked,
	}

	fwdTestTargetRequestTypedData := signer.TypedData{
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
			"from":  forwarederTestTargetRequest.From.String(),
			"to":    forwarederTestTargetRequest.To.String(),
			"value": forwarederTestTargetRequest.Value.String(),
			"gas":   forwarederTestTargetRequest.Gas.String(),
			"nonce": forwarederTestTargetRequest.Nonce.String(),
			"data":  testTargetSetUintIPacked,
		},
	}

	addr, _ := common.NewMixedcaseAddressFromString(signerAddress.String())
	typedDataSignature, _, err := common2.SignTypedData(*signerKey, *addr, fwdTestTargetRequestTypedData)
	require.NoError(t, err)

	testTargetDomainSeparator, err := fwdTestTargetRequestTypedData.HashStruct("EIP712Domain", fwdTestTargetRequestTypedData.Domain.Map())
	require.NoError(t, err)

	testTargetDsep, _ := common2.FromHex(testTargetDomainSeparator.String())

	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", "address from,address to,uint256 value,uint256 gas,uint256 nonce,bytes data")

	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	require.NoError(t, err)

	_, err = forwarder.Execute(owner, *forwarederTestTargetRequest, testTargetDsep, reqTypeHashBytes32, nil, typedDataSignature)

	require.NoError(t, err)
	backend.Commit()

	t.Log(setValue)
	result, err := testTarget.StringData(&bind.CallOpts{})
	require.Equal(t, setValue, result)
	t.Log(result)

}
