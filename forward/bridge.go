package forward

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

// BridgeRequestV2 receiveRequestV2(bytes32 reqId, bytes b, address receiveSide, address bridgeFrom)
func BridgeRequestV2(gsnCaller GsnCaller, chainId *big.Int, signer *ecdsa.PrivateKey, contractAddress common.Address, reqId [32]byte, b []byte, receiveSide common.Address, bridgeFrom [32]byte) (txHash common.Hash, err error) {
	contractABI, err := abi.JSON(strings.NewReader(wrappers.BridgeABI))
	if err != nil {
		return common.Hash{}, fmt.Errorf("could not parse ABI: %w", err)
	}

	fRequest, err := contractABI.Pack("receiveRequestV2", reqId, b, receiveSide, bridgeFrom)
	if err != nil {

		return
	}

	forwarder, err := gsnCaller.GetForwarder(chainId)
	if err != nil {

		return
	}

	forwarderAddress, err := gsnCaller.GetForwarderAddress(chainId)
	if err != nil {

		return
	}

	logrus.Infof("forwarderAddress: %s", forwarderAddress.String())
	logrus.Infof("nodeRegistryAddress: %s", contractAddress.String())

	signerAddress := crypto.PubkeyToAddress(signer.PublicKey)

	logrus.Infof("ownerAddress: %s", signerAddress.String())

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {

		return
	}

	req := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    contractAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  fRequest,
	}

	typedData, err := NewForwardRequestTypedData(
		req,
		forwarderAddress.String(),
		wrappers.BridgeABI,
		"receiveRequestV2", reqId, b, receiveSide, bridgeFrom)
	if err != nil {

		return
	}

	typedDataSignature, _, err := NewSignature(typedData, signer)
	if err != nil {

		return
	}

	domainSeparatorHash, err := NewDomainSeparatorHash(typedData)
	if err != nil {

		return
	}

	genericParams, err := forwarder.GENERICPARAMS(&bind.CallOpts{})
	if err != nil {

		return
	}

	reqTypeHash, err := NewRequestTypeHash(genericParams)
	if err != nil {

		return
	}

	return gsnCaller.Execute(chainId, *req, domainSeparatorHash, reqTypeHash, nil, typedDataSignature)
}
