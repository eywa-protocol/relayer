package forward

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

// NodeListAddNode todo: replace to NodeRegistry.CreateNode
func NodeListAddNode(gsnCaller GsnCaller, chainId *big.Int, signer *ecdsa.PrivateKey, nodeListAddress, nodeListIdAddress common.Address, blsPub string) (txId string, err error) {

	forwarder, err := gsnCaller.GetForwarder(chainId)
	if err != nil {

		return
	}

	forwarderAddress, err := gsnCaller.GetForwarderAddress(chainId)
	if err != nil {

		return
	}

	logrus.Infof("forwarderAddress: %s", forwarderAddress.String())
	logrus.Infof("nodeListAddress: %s", nodeListAddress.String())
	logrus.Infof("nodeListIdAddress: %s", nodeListIdAddress.String())

	signerAddress := crypto.PubkeyToAddress(signer.PublicKey)

	logrus.Infof("signerAddress: %s", signerAddress.String())

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {

		return
	}

	req := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    nodeListAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
	}

	typedData, err := NewForwardRequestTypedData(
		req,
		forwarderAddress.String(),
		wrappers.NodeListABI,
		"addNode", signerAddress, nodeListIdAddress, blsPub)
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

// func NodeRegistryCreateNode(gsnCaller GsnCaller, chainId *big.Int, signer *ecdsa.PrivateKey, nodeRegistryAddress, nodeListIdAddress common.Address, blsPub string) (txId string, err error) {
//
// }
