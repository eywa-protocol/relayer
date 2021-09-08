package networks

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/helpers"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var node *bridge.Node
var err error

var qwe big.Int
var testData *big.Int

func init() {
	err = config.LoadBridgeConfig("../../.data/bridge.yaml")
	if err != nil {
		logrus.Fatal(err)
	}
	node, err = bridge.NewNodeWithClients(context.Background())
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
	logrus.Info("sending to contract ", testData)
	clientFrom := node.Clients[chainidFrom.String()]
	require.NoError(t, err)
	clientTo := node.Clients[chainIdTo.String()]
	require.NoError(t, err)
	dexPoolFrom := node.Clients[chainidFrom.String()].ChainCfg.DexPoolAddress
	dexPoolTo := node.Clients[chainIdTo.String()].ChainCfg.DexPoolAddress
	bridgeTo := node.Clients[chainIdTo.String()].ChainCfg.BridgeAddress
	pKeyFrom := clientFrom.EcdsaKey
	require.NoError(t, err)
	logrus.Print("(dexPoolFrom, clientFrom.EthClient)", dexPoolFrom, clientFrom.EthClient)

	var tx *types.Transaction
	if clientFrom.EthClient != nil {
		txOptsFrom := common2.CustomAuth(clientFrom.EthClient, pKeyFrom)
		dexPoolFromContract, err := wrappers.NewMockDexPool(dexPoolFrom, clientFrom.EthClient)
		require.NoError(t, err)

		tx, err = dexPoolFromContract.SendRequestTestV2(txOptsFrom,
			testData,
			dexPoolTo,
			bridgeTo,
			chainIdTo)

		require.NoError(t, err)
		logrus.Print(tx.Hash())
	} else {
		t.Fatal("eth client from not initialized")
	}

	if clientTo.EthClient != nil {
		dexPoolToContract, err := wrappers.NewMockDexPool(dexPoolTo, clientTo.EthClient)
		require.NoError(t, err)

		status, recaipt := helpers.WaitForBlockCompletation(clientFrom.EthClient, tx.Hash().String())
		logrus.Print(recaipt.Logs)
		logrus.Print(status)
		time.Sleep(20 * time.Second)
		res, err := dexPoolToContract.TestData(&bind.CallOpts{})
		require.NoError(t, err)
		require.Equal(t, testData, res)
		logrus.Print(res)
	} else {
		t.Fatal("eth client to not initialized")
	}

}
