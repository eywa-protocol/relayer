package contracts

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains/eth"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Registry interface {
	NodeRegistry() *wrappers.NodeRegistry
	NodeRegistryFilterer() *wrappers.NodeRegistryFilterer
	NodeRegistrySession(privateKey *ecdsa.PrivateKey) (wrappers.NodeRegistrySession, error)
	NodeRegistryAddress() common.Address
	NodeRegistryPublicKeys() (publicKeys []bls.PublicKey, err error)
	NodeRegistryNodeExists(nodeIdAddress common.Address) bool
}

func NewRegistry(address common.Address, client eth.Client) (*registry, error) {

	if nodeRegistry, err := wrappers.NewNodeRegistry(address, client); err != nil {
		err = fmt.Errorf("init node registry [%s] error: %w", address, err)

		return nil, err
	} else if nodeRegistryFilterer, err := wrappers.NewNodeRegistryFilterer(address, client); err != nil {
		err = fmt.Errorf("init node registry filter [%s] error: %w", address, err)

		return nil, err
	} else {

		return &registry{
			address:            address,
			nodeRegistry:       nodeRegistry,
			nodeRegistryFilter: nodeRegistryFilterer,
		}, nil
	}

}

type registry struct {
	client             eth.Client
	address            common.Address
	nodeRegistry       *wrappers.NodeRegistry
	nodeRegistryFilter *wrappers.NodeRegistryFilterer
}

func (n registry) NodeRegistry() *wrappers.NodeRegistry {

	return n.nodeRegistry
}

func (n registry) NodeRegistrySession(privateKey *ecdsa.PrivateKey) (wrappers.NodeRegistrySession, error) {

	if txOpts, err := n.client.CallOpt(privateKey); err != nil {

		return wrappers.NodeRegistrySession{}, fmt.Errorf("get registry call option error %w", err)
	} else {
		return wrappers.NodeRegistrySession{
			Contract:     n.nodeRegistry,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
		}, nil
	}
}

func (n registry) NodeRegistryFilterer() *wrappers.NodeRegistryFilterer {

	return n.nodeRegistryFilter
}

func (n registry) NodeRegistryAddress() common.Address {

	return n.address
}

func (n registry) NodeRegistryPublicKeys() (publicKeys []bls.PublicKey, err error) {
	publicKeys = make([]bls.PublicKey, 0)
	nodes, err := n.nodeRegistry.GetNodes(&bind.CallOpts{})
	if err != nil {

		return nil, err
	}
	for _, node := range nodes {
		p, err := bls.UnmarshalPublicKey([]byte(node.BlsPubKey))
		if err != nil {
			panic(err)
		}
		publicKeys = append(publicKeys, p)
	}
	return
}

func (n registry) NodeRegistryNodeExists(nodeIdAddress common.Address) bool {
	node, err := n.nodeRegistry.GetNode(&bind.CallOpts{}, nodeIdAddress)
	if err != nil || node.Owner == common.HexToAddress("0") {
		return false
	}
	return true

}
