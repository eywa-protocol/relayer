package test

import (
	"crypto/ecdsa"
	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/helpers"
	"github.com/digiu-ai/p2p-bridge/node/bridge"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"
)

var node *bridge.Node
var err error

var pool_1 common.Address
var pool_2 common.Address
var pool_3 common.Address

var c1 *ethclient.Client
var pk1 *ecdsa.PrivateKey
var c2 *ethclient.Client
var pk2 *ecdsa.PrivateKey

var c3 *ethclient.Client
var pk3 *ecdsa.PrivateKey

var bridgeAddress_1 common.Address
var bridgeAddress_2 common.Address
var bridgeAddress_3 common.Address

var random int

func init() {
	os.Setenv("ECDSA_KEY_1", "0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7")
	os.Setenv("ECDSA_KEY_2", "0x0da632a0af66bc9748f4fe4e8261facffbaef084ae1c591b1d30889622975735")
	os.Setenv("ECDSA_KEY_3", "0xd9a6a69861996d014ad6cf315a0d0c6a5dedaeb77126d2d230e09e5633e8544d")
	node, err = bridge.NewNodeWithClients("")
	pool_1 = common.HexToAddress("0xcfe59d5Bb0A64BFB76D018fDa861941Fb4A5F0E5")
	pool_2 = common.HexToAddress("0xBe4982cA394c5E30183C2C177B28e6A66cF39232")
	pool_3 = common.HexToAddress("0x17908Bd76a3F53310b1eB616B91A12aAa7e759cA")
	bridgeAddress_1 = common.HexToAddress("0x58182e36006a201C8C3f02B2010ba9EE2863626d")
	bridgeAddress_2 = common.HexToAddress("0x3486a8889D92768a078feaB12608ff3a6e186AC3")
	bridgeAddress_3 = common.HexToAddress("0x12FC67a4c837d2dcC53471A025b4061e11472ea8")

}

func Test_SendRequestV2_FromRinkebyToMumbai(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(4), big.NewInt(80001), big.NewInt(rand.Int63()), pool_1, pool_3, bridgeAddress_3)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToRinkeby(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(80001), big.NewInt(4), big.NewInt(rand.Int63()), pool_3, pool_1, bridgeAddress_1)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromRinkebyToBsc(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(4), big.NewInt(97), big.NewInt(rand.Int63()), pool_1, pool_2, bridgeAddress_2)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToRinkeby(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(97), big.NewInt(4), big.NewInt(rand.Int63()), pool_2, pool_1, bridgeAddress_1)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToMumbai(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(97), big.NewInt(80001), big.NewInt(rand.Int63()), pool_2, pool_3, bridgeAddress_3)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToBsc(t *testing.T) {
	err = SendRequestV2FromChainToChain(big.NewInt(80001), big.NewInt(97), big.NewInt(rand.Int63()), pool_3, pool_2, bridgeAddress_2)
	require.NoError(t, err)
}

func SendRequestV2FromChainToChain(chainidFrom, chainIdTo, testData *big.Int, dexPoolFrom, dexPoolTo, bridgeTo common.Address) (err error) {
	logrus.Info("sending to contract ", testData)
	clientFrom, err := node.GetNodeClientByChainId(chainidFrom)
	if err != nil {
		return
	}
	clientTo, err := node.GetNodeClientByChainId(chainIdTo)
	if err != nil {
		return
	}

	pKeyFrom, err := common2.ToECDSAFromHex(clientFrom.ECDSA_KEY)
	if err != nil {
		return
	}
	txOptsFrom := common2.CustomAuth(clientFrom.EthClient, pKeyFrom)
	dexPoolFromContract, err := wrappers.NewMockDexPool(dexPoolFrom, clientFrom.EthClient)

	dexPoolToContract, err := wrappers.NewMockDexPool(dexPoolTo, clientTo.EthClient)

	tx, err := dexPoolFromContract.SendRequestTestV2(txOptsFrom,
		testData,
		dexPoolTo,
		bridgeTo,
		chainIdTo)
	if err != nil {
		return
	}
	logrus.Print(tx.Hash())
	status, recaipt := helpers.WaitForBlockCompletation(clientFrom.EthClient, tx.Hash().String())
	logrus.Print(recaipt.Logs)
	logrus.Print(status)
	time.Sleep(20 * time.Second)
	res, err := dexPoolToContract.TestData(&bind.CallOpts{})
	if err != nil {
		return
	}
	logrus.Print(res)
	return
}

//func Test_Deploy_DexPool_Mumbai(t *testing.T) {
//	txOpts1 := common2.CustomAuth(c3, pk3)
//	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts1, c3, bridgeAddress_3)
//	require.NoError(t, err)
//	require.NotNil(t, tx)
//	require.NotEqual(t, common.HexToAddress("0"), addr)
//	require.NotNil(t, dexPool)
//	logrus.Print(addr, tx.Hash(), dexPool, err)
//}

//func Test_Deploy_DExPoll_Rinkeby(t *testing.T) {
//
//	txOpts1 := common2.CustomAuth(c1, pk1)
//	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts1, c1, bridgeAddress_1)
//	require.NoError(t, err)
//	require.NotEqual(t, common.HexToAddress("0"), addr)
//	require.NotNil(t, dexPool)
//	logrus.Print(addr, tx.Hash(), dexPool)
//}

//func Test_UpdateDexBind_Rinkeby(t *testing.T) {
//	pk := pk1
//	txOpts1 := common2.CustomAuth(c1, pk)
//	br, err := wrappers.NewBridge(bridgeAddress_1, c1)
//	require.NoError(t, err)
//	tx, err := br.UpdateDexBind(txOpts1, pool_1, true)
//	require.NoError(t, err)
//	require.NotNil(t, tx)
//	logrus.Print(tx.Hash())
//	recaipt, err := c1.TransactionReceipt(context.Background(), tx.Hash())
//	require.NoError(t, err)
//	require.NotEqual(t, uint64(0), recaipt.Status)
//
//}
//
//func Test_Deploy_DExPoll_BSCTestnet(t *testing.T) {
//	pk := pk2
//	txOpts2 := common2.CustomAuth(c2, pk)
//	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts2, c2, bridgeAddress_2)
//	require.NoError(t, err)
//	logrus.Print(addr, tx.Hash(), dexPool)
//}
//
//func Test_UpdateDexBind_BSCTestnet(t *testing.T) {
//	pk := pk2
//	txOpts1 := common2.CustomAuth(c2, pk)
//	br, err := wrappers.NewBridge(bridgeAddress_2, c2)
//	require.NoError(t, err)
//	tx, err := br.UpdateDexBind(txOpts1, common.HexToAddress("0x8EFA4F17704D0e8c8678fD67C00a67cec92Cab6e"), true)
//	require.NoError(t, err)
//	require.NotNil(t, tx)
//	logrus.Print(tx.Hash())
//	recaipt, err := helpers.WaitTransaction(c1, tx)
//	require.NoError(t, err)
//	require.NotEqual(t, uint64(0), recaipt.Status)
//}
