package test

import (
	"context"
	"crypto/ecdsa"
	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/helpers"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

var c1 *ethclient.Client
var pk1 *ecdsa.PrivateKey
var c2 *ethclient.Client
var pk2 *ecdsa.PrivateKey

var c3 *ethclient.Client
var pk3 *ecdsa.PrivateKey

var err error
var bridgeAddress_1 common.Address
var bridgeAddress_2 common.Address
var bridgeAddress_3 common.Address

var pool_1 common.Address
var pool_2 common.Address
var pool_3 common.Address

var random int

func init() {
	//oldArgs := os.Args
	//defer func() { os.Args = oldArgs }()
	//
	//fmt.Println(len(os.Args), os.Args)
	//if os.Args[5] != "" {
	//	random, _ = strconv.Atoi(os.Args[5])
	//	logrus.Print("random ",random)
	//}
	c1, err = ethclient.Dial("wss://rinkeby.infura.io/ws/v3/ab95bf9f6dd743e6a8526579b76fe358")
	if err != nil {
		logrus.Fatal(err)
	}
	c2, err = ethclient.Dial("https://data-seed-prebsc-1-s1.binance.org:8545/")
	if err != nil {
		logrus.Fatal(err)
	}
	c3, err = ethclient.Dial("wss://rpc-mumbai.maticvigil.com/ws")
	if err != nil {
		logrus.Fatal(err)
	}

	pk1, err = common2.ToECDSAFromHex("")
	if err != nil {
		logrus.Fatal(err)
	}
	pk2, err = common2.ToECDSAFromHex("")
	if err != nil {
		logrus.Fatal(err)
	}

	pk3, err = common2.ToECDSAFromHex("")
	if err != nil {
		logrus.Fatal(err)
	}
	bridgeAddress_1 = common.HexToAddress("0xa615d60b9D6F080cD47BbB529D088B8c70d8fbd8")
	bridgeAddress_2 = common.HexToAddress("0xfA9DC1525f944440fb1513aA2498422680615a7B")
	bridgeAddress_3 = common.HexToAddress("0xC38a567BC4c59Cc5306B3EaC2A4Df8AC85B243F3")
	pool_1 = common.HexToAddress("0x7f843D9eC365610E6F001a07C5C89ef6226dbdCE")
	pool_2 = common.HexToAddress("0x8EFA4F17704D0e8c8678fD67C00a67cec92Cab6e")
	pool_3 = common.HexToAddress("0x2b0B411DFa42B36649d6D35bba383645C1147067")
}

//
//func Test_Print_DExPoll_Methods(t *testing.T) {
//	//common2.PrintMethods(wrappers.BridgeABI)
//}
//

func Test_Deploy_DexPool_Mumbai(t *testing.T) {
	txOpts1 := common2.CustomAuth(c3, pk3)
	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts1, c3, bridgeAddress_3)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.NotEqual(t, common.HexToAddress("0"), addr)
	require.NotNil(t, dexPool)
	logrus.Print(addr, tx.Hash(), dexPool, err)
}

func Test_Deploy_DExPoll_Rinkeby(t *testing.T) {

	txOpts1 := common2.CustomAuth(c1, pk1)
	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts1, c1, bridgeAddress_1)
	require.NoError(t, err)
	require.NotEqual(t, common.HexToAddress("0"), addr)
	require.NotNil(t, dexPool)
	logrus.Print(addr, tx.Hash(), dexPool)
}

func Test_UpdateDexBind_Rinkeby(t *testing.T) {
	pk := pk1
	txOpts1 := common2.CustomAuth(c1, pk)
	br, err := wrappers.NewBridge(bridgeAddress_1, c1)
	require.NoError(t, err)
	tx, err := br.UpdateDexBind(txOpts1, pool_1, true)
	require.NoError(t, err)
	require.NotNil(t, tx)
	logrus.Print(tx.Hash())
	recaipt, err := c1.TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.NotEqual(t, uint64(0), recaipt.Status)

}

func Test_DeplDExPoll_SendRequestV2_Rinkeby(t *testing.T) {
	pk := pk1
	txOpts1 := common2.CustomAuth(c1, pk)
	dexPool, err := wrappers.NewMockDexPool(pool_1, c1)
	require.NoError(t, err)
	dexPool2, err := wrappers.NewMockDexPool(pool_2, c2)
	require.NoError(t, err)
	//bridge2, err := wrappers.NewBridge(bridgeAddress_2, c2)
	//require.NoError(t, err)

	//NewBridge
	logrus.Print(dexPool)
	//random := rand.Uint64()
	//test := big.NewInt(int64(random))
	test := big.NewInt(199)
	tx, err := dexPool.SendRequestTestV2(txOpts1,
		test,
		pool_2,
		bridgeAddress_2,
		big.NewInt(97))

	require.NoError(t, err)
	require.NotNil(t, tx)
	logrus.Print(tx.Hash())
	recaipt, err := helpers.WaitTransaction(c1, tx)
	require.NoError(t, err)
	require.NotEqual(t, uint64(0), recaipt.Status)
	logrus.Print(recaipt.Logs)

	time.Sleep(20 * time.Second)

	res, err := dexPool2.TestData(&bind.CallOpts{})
	require.NoError(t, err)

	require.Equal(t, test, res)
}

func Test_Deploy_DExPoll_BSCTestnet(t *testing.T) {
	pk := pk2
	txOpts2 := common2.CustomAuth(c2, pk)
	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts2, c2, bridgeAddress_2)
	require.NoError(t, err)
	logrus.Print(addr, tx.Hash(), dexPool)
	//=== RUN   Test_Deploy_DExPoll_BSCTestnet
	//time="2021-06-22T21:58:10+03:00" level=info msg="0x8EFA4F17704D0e8c8678fD67C00a67cec92Cab6e 0x51ed3b91dde8c78edfb0eb61e48d0bf5fec256342dcbb947d18e31229a101850 &{{0xc00045e500} {0xc00045e500} {0xc00045e500}}"
	//--- PASS: Test_Deploy_DExPoll_BSCTestnet (2.40s)
	//PASS
}

func Test_UpdateDexBind_BSCTestnet(t *testing.T) {
	pk := pk2
	txOpts1 := common2.CustomAuth(c2, pk)
	br, err := wrappers.NewBridge(bridgeAddress_2, c2)
	require.NoError(t, err)
	tx, err := br.UpdateDexBind(txOpts1, common.HexToAddress("0x8EFA4F17704D0e8c8678fD67C00a67cec92Cab6e"), true)
	require.NoError(t, err)
	require.NotNil(t, tx)
	logrus.Print(tx.Hash())
	recaipt, err := helpers.WaitTransaction(c1, tx)
	require.NoError(t, err)
	require.NotEqual(t, uint64(0), recaipt.Status)

}

func Test_SendRequestV2_Bsctestnet(t *testing.T) {
	pk, err := common2.ToECDSAFromHex("0xbb57ba69119c43d5f9f7f2a2799c438ed1857338a885b304ce94b322aa428ab9")
	require.NoError(t, err)
	txOpts1 := common2.CustomAuth(c1, pk)
	addr, tx, dexPool, err := wrappers.DeployMockDexPool(txOpts1, c1, bridgeAddress_1)
	require.NoError(t, err)
	logrus.Print(addr, tx.Hash(), dexPool)
	/*
		=== RUN   TestDeplDExPollRBSCTestnet
		time="2021-06-22T20:12:36+03:00" level=info msg="0xD2F6946f773156023caEdCC300139D450e27aD41 0x13718f3ff45cb83dc4d9aac4afa72ea3ff8f20585972901a713e0c2b632ea3db &{{0xc00046e500} {0xc00046e500} {0xc00046e500}}"
		--- PASS: TestDeplDExPollRBSCTestnet (4.37s)
		PASS
	*/
}
