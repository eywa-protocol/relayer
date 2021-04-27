package common

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/linkpoolio/bridges"
	"github.com/sirupsen/logrus"
	"strings"
)

func Connect(string2 string) (*ethclient.Client, error) {
	return ethclient.Dial(string2)
}

func Health(helper *bridges.Helper, rpcUrl string) (out *Output, err error) {
	out = &Output{}
	client, err := Connect(rpcUrl)
	if err != nil {
		return
	}
	block, err := client.BlockNumber(context.Background())
	if err != nil {
		return
	}
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	out.ChainId = fmt.Sprintf("%d", chainId)
	out.BlockNum = fmt.Sprintf("%d", block)
	return
}

func HealthFirst(helper *bridges.Helper) (out *Output, err error) {
	return Health(helper, config.Config.NETWORK_RPC_1)
}

func HealthSecond(helper *bridges.Helper) (*Output, error) {
	return Health(helper, config.Config.NETWORK_RPC_2)
}

func ToECDSAFromHex(hexString string) (pk *ecdsa.PrivateKey, err error) {
	pk, err = crypto.HexToECDSA(strings.TrimPrefix(hexString, "0x"))
	return
}

/*func SetMockPoolTestRequestV2(helper *bridges.Helper) (o *Output, err error) {
	o = &Output{}

	reqId := helper.GetIntParam("id")
	client1, err := Connect(config.Config.NETWORK_RPC_1)
	if err != nil {
		return

	}

	pKey1, err := ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	txOpts1 := bind.NewKeyedTransactor(pKey1)

	mockBridgeContract1, err := wrappers.NewBridge(common.HexToAddress(config.Config.BRIDGE_ADDRESS_NETWORK1), client1)
	if err != nil {
		return
	}

	tx, err := mockBridgeContract1.TransmitRequestV2(txOpts1, big.NewInt(reqId), common.HexToAddress(config.Config.TOKENPOOL_ADDRESS_2))

	if err != nil {
		return
	}

	logrus.Printf("TX HASH %x", tx.Hash())
	o.ChainId = fmt.Sprintf("%s", tx.ChainId())
	o.TxHash = tx.Hash().Hex()
	return
}*/

func RegisterNode(client *ethclient.Client, nodeListContractAddress common.Address, nodeWallet common.Address, p2pAddress []byte, pubKey []byte) (err error) {
	logrus.Printf("REGISTERING NODE sender:%v", nodeWallet)
	pKey1, err := ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	txOpts1 := bind.NewKeyedTransactor(pKey1)

	nodeListContract1, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	tx, err := nodeListContract1.AddNode(txOpts1, nodeWallet, p2pAddress, pubKey, true)

	if err != nil {
		return
	}

	logrus.Printf("TX HASH %x", tx.Hash().Hex())
	return
}

func ChainlinkData(helper *bridges.Helper) (o *Output, err error) {
	o = &Output{}
	fmt.Print(helper.Data)
	o.Data2 = *helper.Data
	return
}

/*func GetP2PBootstrapPeerId(helper *bridges.Helper) (o *Output, err error) {
	o = &Output{}
	reqId := helper.GetIntParam("id")
	client1, err := Connect(config.Config.NETWORK_RPC_1)
	if err != nil {
		return

	}

	pKey1, err := ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	txOpts1 := bind.NewKeyedTransactor(pKey1)

	mockDexPoolContract1, err := wrappers.NewMockDexPool(common.HexToAddress(config.Config.TOKENPOOL_ADDRESS_1), client1)
	if err != nil {
		return
	}

	tx, err := mockDexPoolContract1.SendRequestTest(txOpts1, big.NewInt(reqId), common.HexToAddress(config.Config.TOKENPOOL_ADDRESS_2))

	if err != nil {
		return
	}

	logrus.Printf("TX HASH %x", tx.Hash())
	o.ChainId = fmt.Sprintf("%s", tx.ChainId())
	o.TxHash = tx.Hash().Hex()
	return
}*/
