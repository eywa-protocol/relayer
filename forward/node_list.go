package forward

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

func NodeRegistryCreateNode(gsnCaller GsnCaller, chainId *big.Int, signer *ecdsa.PrivateKey, nodeRegistryAddress common.Address, node wrappers.NodeRegistryNode, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (txHash common.Hash, err error) {
	nodeRegistryABI := common2.MustGetABI(wrappers.NodeRegistryABI)

	frequest, err := nodeRegistryABI.Pack("createRelayer", node, deadline, v, r, s)
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
	logrus.Infof("nodeRegistryAddress: %s", nodeRegistryAddress.String())

	signerAddress := crypto.PubkeyToAddress(signer.PublicKey)

	logrus.Infof("ownerAddress: %s", signerAddress.String())

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, signerAddress)
	if err != nil {

		return
	}

	req := &wrappers.IForwarderForwardRequest{
		From:  signerAddress,
		To:    nodeRegistryAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  frequest,
	}

	typedData, err := NewForwardRequestTypedData(
		req,
		forwarderAddress.String(),
		wrappers.NodeRegistryABI,
		"createRelayer", node, deadline, v, r, s)
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
