package common

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/digiu-ai/p2p-bridge/config"
	wrappers "github.com/digiu-ai/wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/linkpoolio/bridges"
	"github.com/sirupsen/logrus"
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

func GetChainIdByurl(rpcUrl string) (chainId *big.Int, err error) {
	client, err := Connect(rpcUrl)
	if err != nil {
		return
	}
	chainId, err = client.ChainID(context.Background())
	if err != nil {
		return
	}
	return
}

func GetClientByChainId(chainIdFromClient *big.Int) (client *ethclient.Client, err error) {
	client, err = Connect(config.Config.NETWORK_RPC_1)
	if err != nil {
		return
	}
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return
	}
	if chainId.Cmp(chainIdFromClient) == 0 {
		return
	}
	logrus.Tracef("CHAIND 1 %d chainIdFromClient %d ", chainId, chainIdFromClient)
	client, err = Connect(config.Config.NETWORK_RPC_2)
	if err != nil {
		return
	}
	chainId, err = client.ChainID(context.Background())
	if err != nil {
		return
	}
	if chainId.Cmp(chainIdFromClient) == 0 {
		return
	}
	logrus.Tracef("CHAIND 2 %d chainIdFromClient %d ", chainId, chainIdFromClient)
	client, err = Connect(config.Config.NETWORK_RPC_3)
	if err != nil {
		return
	}
	chainId, err = client.ChainID(context.Background())
	if err != nil {
		return
	}
	if chainId.Cmp(chainIdFromClient) == 0 {
		return
	}
	logrus.Tracef("CHAIND 3 %d chainIdFromClient %d ", chainId, chainIdFromClient)
	return nil, errors.New(fmt.Sprintf("not found rpcurl for chainID %d", chainIdFromClient))
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

func CreateNode(duration time.Duration, creaateNode func(time.Time) bool) chan bool {
	ticker := time.NewTicker(duration)
	stop := make(chan bool, 1)

	go func() {
		defer log.Println("CreateNode ticker stopped")
		for {
			select {
			case time := <-ticker.C:
				if !creaateNode(time) {
					stop <- true
				}
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func CreateNodeWithTicker(ctx context.Context, c ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	for {

		logrus.Info("WaitTransaction ", "txHash", txHash)

		receipt, err := c.TransactionReceipt(ctx, txHash)
		if receipt != nil {
			return receipt, nil
		}
		if err != nil {
			logrus.Error("Receipt retrieval failed", "err", err)
		} else {
			logrus.Trace("Transaction not yet mined")
		}
		// Wait for the next round.
		select {

		case <-ctx.Done():

			return nil, ctx.Err()

		case <-queryTicker.C:

		}
	}
}

func RegisterNode(client *ethclient.Client, pk *ecdsa.PrivateKey, nodeListContractAddress common.Address, p2pAddress []byte, blsPubkey []byte, blsAddr common.Address) (err error) {
	logrus.Infof("Register Node %x", blsAddr)
	fromAddress := crypto.PubkeyToAddress(*(pk.Public().(*ecdsa.PublicKey)))
	nodeListContract1, err := wrappers.NewNodeList(nodeListContractAddress, client)
	res, err := nodeListContract1.NodeExists(&bind.CallOpts{}, blsAddr)
	if res {
		logrus.Infof("node %x allready exists", blsAddr)
	} else {
		logrus.Infof("creating peer")
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		mychannel := make(chan bool)
		for {
			select {
			case <-ticker.C:
				txOpts1 := CustomAuth(client, pk)
				tx, err := nodeListContract1.AddNode(txOpts1, fromAddress, p2pAddress, blsAddr, blsPubkey, true)
				if tx != nil {
					if created, _ := nodeListContract1.NodeExists(&bind.CallOpts{}, blsAddr); created {
						logrus.Infof("Added Node: %x blsAddR: %v txhash %x", fromAddress, blsAddr, tx.Hash())
						ticker.Stop()
						return nil
					}

				}
				if err != nil {
					logrus.Errorf("AddNode ERROR: %v", err)
				}
			case <-mychannel:
				return
			}
		}
		time.Sleep(15 * time.Second)
		ticker.Stop()
		mychannel <- true
	}
	return
}

func GetNodesFromContract(client *ethclient.Client, nodeListContractAddress common.Address) (nodes []wrappers.NodeListNode, err error) {
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
		logrus.Tracef("node ID %d p2pAddress %v  BlsPointAddr %v", node1.NodeId, string(node1.P2pAddress), node1.BlsPointAddr)
	}

}

func GetNode(client *ethclient.Client, nodeListContractAddress common.Address, nodeBLSAddr common.Address) (node wrappers.NodeListNode, err error) {
	nodeList, err := wrappers.NewNodeList(nodeListContractAddress, client)
	if err != nil {
		return
	}

	if exist, _ := nodeList.NodeExists(&bind.CallOpts{}, nodeBLSAddr); !exist {
		err = errors.New(fmt.Sprintf("NODE DOES NOT EXIST %v", nodeBLSAddr))
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

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
