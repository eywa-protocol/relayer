package common

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/helpers"
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

func RegisterNode(client *ethclient.Client, pk *ecdsa.PrivateKey, nodeListContractAddress common.Address, nodeWallet common.Address, p2pAddress []byte, blsPubkey []byte, blsAddr common.Address) (err error) {
	txOpts1 := bind.NewKeyedTransactor(pk)
	nodeListContract1, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	res, err := nodeListContract1.NodeExists(&bind.CallOpts{}, blsAddr)
	if err != nil {
		return
	}

	if !res {
		logrus.Printf("client: %v REGISTERING node %v in contract %v NODE sender:%v with PK %v blsAddress %v ", client, string(p2pAddress), nodeListContractAddress, nodeWallet, pk, blsAddr)
		tx, err := nodeListContract1.AddNode(txOpts1, nodeWallet, p2pAddress, blsAddr, blsPubkey, true)
		if err != nil {
			return err
		}
		logrus.Printf("TX HASH %x", tx.Hash().Hex())
		receipt, err := helpers.WaitTransaction(client, tx)
		if err != nil {
			return err
		}
		if receipt == nil {
			return errors.New(fmt.Sprintf("AddNode Failed %v %v %v %v %v", txOpts1, nodeWallet, p2pAddress, blsAddr, blsPubkey))
		}
	}
	return
}

func GetNodesFromContract(client *ethclient.Client, nodeListContractAddress common.Address) (nodes []wrappers.NodeListNode, err error) {
	logrus.Printf("GetNodesFromContract: %v", nodeListContractAddress)
	nodeList, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	nodes, err = nodeList.GetNodes(&bind.CallOpts{})

	if err != nil {
		return
	}

	return
}

func PrintNodes(client *ethclient.Client, nodeListContractAddress common.Address) {
	nodeList, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	nodes, err := nodeList.GetNodes(&bind.CallOpts{})
	if err != nil {
		logrus.Fatal(err)
	}
	for _, node1 := range nodes {
		logrus.Printf("node ID %d p2pAddress %v  BlsPointAddr %v", node1.NodeId, string(node1.P2pAddress), node1.BlsPointAddr)
	}

}

func GetNode(client *ethclient.Client, nodeListContractAddress common.Address, nodeBLSAddr common.Address) (node wrappers.NodeListNode, err error) {
	logrus.Printf("GetNode : %v", nodeListContractAddress)
	nodeList, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	if exist, _ := nodeList.NodeExists(&bind.CallOpts{}, nodeBLSAddr); !exist {
		errMsg := fmt.Sprintf("NODE DOES NOT EXIST %v", nodeBLSAddr)
		err = errors.New(errMsg)
		return
	}
	node, err = nodeList.GetNode(&bind.CallOpts{}, nodeBLSAddr)
	if err != nil {
		return
	}
	return
}

func ChainlinkData(helper *bridges.Helper) (o *Output, err error) {
	o = &Output{}
	fmt.Print(helper.Data)
	o.Data2 = *helper.Data
	return
}
