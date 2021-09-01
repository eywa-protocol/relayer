package local

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/forward"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/encoding"
)

type gsnClientNodeType struct {
	base.Node
	ecdsaPriv *ecdsa.PrivateKey
	priv      crypto.PrivKey
	pub       crypto.PubKey
	blsPriv   kyber.Scalar
	blsPub    kyber.Point
	gsnClient *gsn.Client
}

func NewGsnClientNode(ctx context.Context) (node *gsnClientNodeType, err error) {
	n := &gsnClientNodeType{
		Node: base.Node{
			Ctx: ctx,
		},
	}

	if err := config.LoadBridgeConfig("../../.data/bridge.yaml", true); err != nil {
		logrus.Fatal(err)
	}

	n.ecdsaPriv, err = ecdsa.GenerateKey(crypto2.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	n.priv, n.pub, err = crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}

	suite := pairing.NewSuiteBn256()
	n.blsPriv = suite.Scalar().Pick(suite.RandomStream())
	n.blsPub = suite.Point().Mul(n.blsPriv, nil)

	multiAddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	if err != nil {
		return nil, err
	}

	n.Host, err = libp2p.NewHostFromKey(n.priv, multiAddr)
	if err != nil {
		return nil, err
	}

	n.Dht, err = n.InitDHT(config.Bridge.BootstrapAddrs)
	if err != nil {
		return nil, err
	}

	n.gsnClient, err = gsn.NewClient(ctx, n.Host, n, config.Bridge.TickerInterval)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *gsnClientNodeType) WaitForDiscovery(timeout time.Duration) error {
	return n.gsnClient.WaitForDiscoveryGsn(timeout)
}

func (n *gsnClientNodeType) ExecuteNodeRegistryCreateNode(chainId *big.Int) (txHash common.Hash, err error) {

	chainCfg, err := n.getChainCfg(chainId)
	if err != nil {
		return
	}
	suite := pairing.NewSuiteBn256()
	blsPub, err := encoding.PointToStringHex(suite, n.blsPub)
	if err != nil {
		return
	}
	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))

	fromAddress := common2.AddressFromSecp256k1PrivKey(n.ecdsaPriv)

	ethClient, _, err := chainCfg.GetEthClient("", fromAddress)
	if err != nil {
		err = fmt.Errorf("init eth client [%s] error: %w", chainId.String(), err)
		return
	}

	nodeRegistry, err := wrappers.NewNodeRegistry(chainCfg.NodeRegistryAddress, ethClient)
	if err != nil {
		err = fmt.Errorf("init node registry [%s] error: %w", chainCfg.NodeRegistryAddress, err)
		return
	}

	eywaAddress, _ := nodeRegistry.EYWA(&bind.CallOpts{})
	eywa, err := wrappers.NewERC20Permit(eywaAddress, ethClient)
	if err != nil {
		err = fmt.Errorf("EYWA contract: %w", err)
		return
	}

	fromNonce, _ := eywa.Nonces(&bind.CallOpts{}, fromAddress)
	value, _ := eywa.BalanceOf(&bind.CallOpts{}, fromAddress)

	deadline := big.NewInt(time.Now().Unix() + 100)
	const EywaPermitName = "EYWA"
	const EywaPermitVersion = "1"
	v, r, s := common2.SignErc20Permit(n.ecdsaPriv, EywaPermitName, EywaPermitVersion, chainId,
		eywaAddress, fromAddress, chainCfg.NodeRegistryAddress, value, fromNonce, deadline)

	node := wrappers.NodeRegistryNode{
		Owner:         fromAddress,
		NodeIdAddress: nodeIdAddress,
		Pool:          fromAddress,
		BlsPubKey:     blsPub,
		NodeId:        nodeIdAddress.Hash().Big(),
	}

	return forward.NodeRegistryCreateNode(n.gsnClient, chainId, n.ecdsaPriv, chainCfg.NodeRegistryAddress, node, deadline, v, r, s)
}

func (n *gsnClientNodeType) getChainCfg(chainId *big.Int) (*config.BridgeChain, error) {
	for _, chain := range config.Bridge.Chains {
		if uint64(chain.Id) == chainId.Uint64() {
			return chain, nil
		}
	}

	return nil, fmt.Errorf("invalid chain [%s]", chainId.String())
}

func (n *gsnClientNodeType) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {

	chainCfg, err := n.getChainCfg(chainId)
	if err != nil {
		return nil, err
	}

	ethClient, _, err := chainCfg.GetEthClient("", common2.AddressFromSecp256k1PrivKey(n.ecdsaPriv))
	if err != nil {
		return nil, err
	}

	return wrappers.NewForwarder(chainCfg.ForwarderAddress, ethClient)
}

func (n *gsnClientNodeType) GetForwarderAddress(chainId *big.Int) (common.Address, error) {
	chainCfg, err := n.getChainCfg(chainId)
	if err != nil {
		return common.Address{}, err
	}

	return chainCfg.ForwarderAddress, nil
}

func TestGsnClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n, err := NewGsnClientNode(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = n.WaitForDiscovery(30 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	for _, chainCfg := range config.Bridge.Chains {

		txHash, err := n.ExecuteNodeRegistryCreateNode(big.NewInt(int64(chainCfg.Id)))
		assert.NoError(t, err)
		assert.NotEmpty(t, txHash)
		fmt.Println(txHash)

	}

}
