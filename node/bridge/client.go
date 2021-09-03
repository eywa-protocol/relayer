package bridge

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/forward"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/helpers"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Client struct {
	EthClient            *ethclient.Client
	ChainCfg             *config.BridgeChain
	Bridge               wrappers.BridgeSession
	BridgeFilterer       wrappers.BridgeFilterer
	NodeRegistry         wrappers.NodeRegistrySession
	NodeRegistryFilterer wrappers.NodeRegistryFilterer
	Forwarder            wrappers.Forwarder

	currentUrl string
}

func NewClient(chain *config.BridgeChain, skipUrl string, signerKey *ecdsa.PrivateKey) (client Client, err error) {
	var signerAddress common.Address
	if signerKey != nil {
		signerAddress = common2.AddressFromSecp256k1PrivKey(signerKey)
	} else {
		signerAddress = chain.EcdsaAddress
	}
	c, url, err := chain.GetEthClient(skipUrl, signerAddress)
	if err != nil {
		err = fmt.Errorf("get eth client error: %w", err)
		return
	}
	chain.ChainId, err = c.ChainID(context.Background())
	if err != nil {
		err = fmt.Errorf("get chain id error: %w", err)
		return
	}

	bridge, err := wrappers.NewBridge(chain.BridgeAddress, c)
	if err != nil {
		err = fmt.Errorf("init bridge [%s] error: %w", chain.BridgeAddress, err)
		return
	}

	bridgeFilterer, err := wrappers.NewBridgeFilterer(chain.BridgeAddress, c)
	if err != nil {
		err = fmt.Errorf("init bridge filter [%s] error: %w", chain.BridgeAddress, err)
		return
	}
	nodeRegistry, err := wrappers.NewNodeRegistry(chain.NodeRegistryAddress, c)
	if err != nil {
		err = fmt.Errorf("init node registry [%s] error: %w", chain.NodeRegistryAddress, err)
		return
	}

	nodeListFilterer, err := wrappers.NewNodeRegistryFilterer(chain.NodeRegistryAddress, c)
	if err != nil {
		err = fmt.Errorf("init nodelist filter [%s] error: %w", chain.NodeRegistryAddress, err)
		return
	}

	forwarder, err := wrappers.NewForwarder(chain.ForwarderAddress, c)
	if err != nil {
		err = fmt.Errorf("init forwarder caller [%s] error: %w", chain.BridgeAddress, err)
		return
	}

	var txOpts *bind.TransactOpts
	if signerKey != nil {
		txOpts = common2.CustomAuth(c, signerKey)
	} else {
		txOpts = common2.CustomAuth(c, chain.EcdsaKey)
	}

	return Client{
		EthClient: c,
		ChainCfg:  chain,
		Bridge: wrappers.BridgeSession{
			Contract:     bridge,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
		},
		BridgeFilterer: *bridgeFilterer,
		NodeRegistry: wrappers.NodeRegistrySession{
			Contract:     nodeRegistry,
			CallOpts:     bind.CallOpts{},
			TransactOpts: *txOpts,
		},
		NodeRegistryFilterer: *nodeListFilterer,
		Forwarder:            *forwarder,
		currentUrl:           url,
	}, nil
}

func (c *Client) RegisterNode(gsnClient *gsn.Client, ownerPrivKey *ecdsa.PrivateKey, peerId peer.ID, blsPubkey string) (id *big.Int, relayerPool *common.Address, err error) {
	logrus.Infof("Adding Node %s it's NodeidAddress %x", peerId, common.BytesToAddress([]byte(peerId.String())))
	fromAddress := common2.AddressFromSecp256k1PrivKey(ownerPrivKey)
	nodeIdAsAddress := common.BytesToAddress([]byte(peerId))

	res, err := c.NodeRegistry.NodeExists(nodeIdAsAddress)
	if err != nil {
		err = fmt.Errorf("node not exists nodeIdAddress: %s, client.Id: %s, error: %w",
			nodeIdAsAddress.String(), c.ChainCfg.ChainId.String(), err)
	}
	if res == true {
		logrus.Infof("Node %x allready exists", peerId)
		return
	}

	eywaAddress, _ := c.NodeRegistry.EYWA()
	eywa, err := wrappers.NewERC20Permit(eywaAddress, c.EthClient)
	if err != nil {
		return nil, nil, fmt.Errorf("EYWA contract: %w", err)
	}
	fromNonce, _ := eywa.Nonces(&bind.CallOpts{}, fromAddress)
	value, _ := eywa.BalanceOf(&bind.CallOpts{}, fromAddress)

	deadline := big.NewInt(time.Now().Unix() + 100)
	const EywaPermitName = "EYWA"
	const EywaPermitVersion = "1"
	v, r, s := common2.SignErc20Permit(ownerPrivKey, EywaPermitName, EywaPermitVersion, c.ChainCfg.ChainId,
		eywaAddress, fromAddress, c.ChainCfg.NodeRegistryAddress, value, fromNonce, deadline)

	node := wrappers.NodeRegistryNode{
		Owner:         fromAddress,
		Pool:          common.Address{},
		NodeIdAddress: nodeIdAsAddress,
		BlsPubKey:     blsPubkey,
		NodeId:        big.NewInt(0),
	}

	var txHash common.Hash
	if c.ChainCfg.UseGsn && gsnClient != nil {
		if txHash, err = forward.NodeRegistryCreateNode(gsnClient, c.ChainCfg.ChainId, ownerPrivKey, c.ChainCfg.NodeRegistryAddress, node, deadline, v, r, s); err != nil {
			err = fmt.Errorf("CreateRelayer over gsn chainId %d ERROR: %v", c.ChainCfg.ChainId, err)
			logrus.Error(err)
			return nil, nil, err
		}
	} else {
		tx, err := c.NodeRegistry.CreateRelayer(node, deadline, v, r, s)
		if err != nil {
			err = fmt.Errorf("CreateRelayer chainId %d ERROR: %v", c.ChainCfg.ChainId, err)
			logrus.Error(err)
			return nil, nil, err
		}
		txHash = tx.Hash()
	}

	receipt, err := helpers.WaitTransactionDeadline(c.EthClient, txHash, 100*time.Second)
	if err != nil {

		return nil, nil, fmt.Errorf("WaitTransaction error: %w", err)
	}
	logrus.Infof("recept.Status %d", receipt.Status)

	blockNum := receipt.BlockNumber.Uint64()

	it, err := c.NodeRegistryFilterer.FilterCreatedRelayer(&bind.FilterOpts{Start: blockNum, End: &blockNum},
		[]common.Address{node.NodeIdAddress}, []*big.Int{}, []common.Address{})
	if err != nil {

		return nil, nil, err
	}
	defer func() {
		if err := it.Close(); err != nil {

			logrus.Error(fmt.Errorf("close registry created rellayer iterator error: %w", err))
		}
	}()

	for it.Next() {
		logrus.Info("CreatedRelayer Event", it.Event.NodeIdAddress)
		return it.Event.NodeId, &it.Event.RelayerPool, nil
	}

	return
}
