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

	"github.com/libp2p/go-libp2p-core/peer"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/helpers"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/linkpoolio/bridges"
	"github.com/sirupsen/logrus"
	wrappers "gitlab.digiu.ai/blockchainlaboratory/wrappers"
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

func RegisterNode(client *ethclient.Client, pk *ecdsa.PrivateKey, nodeListContractAddress common.Address, peerId peer.ID, blsPubkey string) (err error) {
	logrus.Infof("Adding Node %s it's NodeidAddress %x", peerId, common.BytesToAddress([]byte(peerId.String())))
	fromAddress := crypto.PubkeyToAddress(*(pk.Public().(*ecdsa.PublicKey)))

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("get chain id error: %w", err)
	}
	nodeListContract1, err := wrappers.NewNodeList(nodeListContractAddress, client)
	res, err := nodeListContract1.NodeExists(&bind.CallOpts{}, common.BytesToAddress([]byte(peerId)))
	if err != nil {
		err = fmt.Errorf("node not exists nodeListContractAddress: %s, client.Id: %s, error: %w",
			nodeListContractAddress.String(), chainId.String(), err)
	}
	if res == true {
		logrus.Infof("Node %x allready exists", peerId)
	} else {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		mychannel := make(chan bool)
		for {
			select {
			case <-ticker.C:
				created, err := nodeListContract1.NodeExists(&bind.CallOpts{}, common.BytesToAddress([]byte(peerId)))
				if err != nil {
					logrus.Errorf("NodeExists: %v", err)
				}
				if created == false {
					txOpts1 := CustomAuth(client, pk)
					tx, err := nodeListContract1.AddNode(txOpts1, fromAddress, common.BytesToAddress([]byte(peerId)), blsPubkey)
					if err != nil {
						chainId, _ := client.ChainID(context.Background())
						logrus.Errorf("AddNode chainId %d ERROR: %v", chainId, err)
						if strings.Contains(err.Error(), "allready exists") || strings.Contains(err.Error(), "gas required exceeds allowance") {
							ticker.Stop()
							return err
						}
					} else {
						recept, _ := helpers.WaitTransaction(client, tx)
						logrus.Print("recept.Status ", recept.Status)
						ticker.Stop()
						return nil

					}

				} else {

					ticker.Stop()
					return nil
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
		logrus.Tracef("node ID %d node.NodeIdAddress %v", node1.NodeId, node1.NodeIdAddress)
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
