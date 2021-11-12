package test

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var node *bridge.Node
var err error
var clientsClose func()

var qwe big.Int
var testData *big.Int

var chainCfg map[uint64]*config.BridgeChain

func initNodeWithClients() {
	err = config.LoadBridgeConfig("../.data/bridge.yaml", false)
	if err != nil {
		logrus.Fatal(err)
	}
	chainCfg = make(map[uint64]*config.BridgeChain, len(config.Bridge.Chains))
	for _, chain := range config.Bridge.Chains {
		chainCfg[chain.Id] = chain
	}
	node, err, clientsClose = bridge.NewNodeWithClients(context.Background(), nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(os.Args) > 5 {
		testData, _ = qwe.SetString(os.Args[5], 10)
	} else {
		testData = qwe.SetUint64(rand.Uint64())
	}
}

func SendRequestV2FromChainToChain(t *testing.T, chainidFrom, chainIdTo, testData *big.Int) {
	initNodeWithClients()
	if clientsClose != nil {
		defer clientsClose()
	}
	logrus.Info("sending to contract ", testData)
	clientFrom, err := node.GetClients().GetEthClient(chainidFrom)
	require.NoError(t, err)
	clientTo, err := node.GetClients().GetEthClient(chainIdTo)
	require.NoError(t, err)
	dexPoolFrom := chainCfg[chainidFrom.Uint64()].DexPoolAddress
	dexPoolTo := chainCfg[chainIdTo.Uint64()].DexPoolAddress
	bridgeTo := chainCfg[chainIdTo.Uint64()].BridgeAddress
	pKeyFrom := chainCfg[chainidFrom.Uint64()].EcdsaKey
	require.NoError(t, err)
	logrus.Print("(dexPoolFrom)", dexPoolFrom)

	txOptsFrom, err := clientFrom.CallOpt(pKeyFrom)
	require.NoError(t, err)

	dexPoolFromContract, err := wrappers.NewMockDexPool(dexPoolFrom, clientFrom)

	dexPoolToContract, err := wrappers.NewMockDexPool(dexPoolTo, clientTo)
	testData = qwe.SetUint64(rand.Uint64())
	require.NotNil(t, testData)
	tx, err := dexPoolFromContract.SendRequestTestV2(txOptsFrom,
		testData,
		dexPoolTo,
		bridgeTo,
		chainIdTo)
	require.NoError(t, err)
	logrus.Print(tx.Hash())
	recaipt, err := clientFrom.WaitTransaction(tx.Hash())
	logrus.Print(recaipt.Logs)
	logrus.Print(recaipt.Status)
	time.Sleep(20 * time.Second)
	res, err := dexPoolToContract.TestData(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, testData, res)
	logrus.Print(res)
}
