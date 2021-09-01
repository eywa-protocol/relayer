package forward

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

const (
	domainType     = "EIP712Domain"
	forwardRequest = "ForwardRequest"
)

type GsnCaller interface {
	GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error)
	GetForwarderAddress(chainId *big.Int) (common.Address, error)
	Execute(chainId *big.Int, req wrappers.IForwarderForwardRequest, domainSeparator [32]byte, requestTypeHash [32]byte, suffixData []byte, sig []byte) (common.Hash, error)
}

func NewForwardRequestTypedData(req *wrappers.IForwarderForwardRequest, forwarderAddress, abiJson, methodName string, args ...interface{}) (*signer.TypedData, error) {

	contractAbi, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return nil, fmt.Errorf("could not parse ABI: %w", err)
	}

	contractAbiPacked, err := contractAbi.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("could not pack abi method: %s on error:  %w", methodName, err)
	}

	req.Data = contractAbiPacked

	return &signer.TypedData{
		Types: signer.Types{
			domainType: []signer.Type{{
				Name: "verifyingContract", Type: "address"}},
			forwardRequest: []signer.Type{
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
			VerifyingContract: forwarderAddress,
		},
		PrimaryType: "ForwardRequest",
		Message: signer.TypedDataMessage{
			"from":  req.From.String(),
			"to":    req.To.String(),
			"value": req.Value.String(),
			"gas":   req.Gas.String(),
			"nonce": req.Nonce.String(),
			"data":  contractAbiPacked,
		},
	}, nil
}

func NewSignature(typedData *signer.TypedData, signer *ecdsa.PrivateKey) (signature hexutil.Bytes, SignDataRequestHash hexutil.Bytes, err error) {

	if addr, err := common.NewMixedcaseAddressFromString(crypto.PubkeyToAddress(signer.PublicKey).String()); err != nil {

		return nil, nil, err
	} else {

		return common2.SignTypedData(*signer, *addr, *typedData)
	}
}

func NewDomainSeparatorHash(typedData *signer.TypedData) (common2.Bytes32, error) {

	if domainSeparator, err := typedData.HashStruct(domainType, typedData.Domain.Map()); err != nil {

		return common2.Bytes32{}, err
	} else {

		return common2.FromHex(domainSeparator.String())
	}

}

func NewRequestTypeHash(genericParams string) (common2.Bytes32, error) {

	forwardRequestType := fmt.Sprintf("%s(%s)", forwardRequest, genericParams)
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))

	if reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash); err != nil {

		return common2.Bytes32{}, err
	} else {

		return reqTypeHashBytes32, nil
	}

}
