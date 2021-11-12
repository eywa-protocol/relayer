package common

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core"
	"github.com/sirupsen/logrus"
)

func Connect(string2 string) (*ethclient.Client, error) {
	return ethclient.Dial(string2)
}

func AddressIsZero(address common.Address) bool {
	return address.String() == common.Address{}.String()
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
			case t := <-ticker.C:
				if !creaateNode(t) {
					stop <- true
				}
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// todo remove commented
// func CreateNodeWithTicker(ctx context.Context, c ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
// 	queryTicker := time.NewTicker(time.Second)
// 	defer queryTicker.Stop()
// 	for {
//
// 		logrus.Info("WaitTransaction ", "txHash", txHash)
//
// 		receipt, err := c.TransactionReceipt(ctx, txHash)
// 		if receipt != nil {
// 			return receipt, nil
// 		}
// 		if err != nil {
// 			logrus.Error("Receipt retrieval failed", "err", err)
// 		} else {
// 			logrus.Trace("Transaction not yet mined")
// 		}
// 		// Wait for the next round.
// 		select {
//
// 		case <-ctx.Done():
//
// 			return nil, ctx.Err()
//
// 		case <-queryTicker.C:
//
// 		}
// 	}
// }

func SignErc20Permit(pk *ecdsa.PrivateKey, name, version string, chainId *big.Int, verifyingContract, owner, spender common.Address, value, nonce, deadline *big.Int) (v uint8, r [32]byte, s [32]byte) {
	data := core.TypedData{
		Types: core.Types{
			"EIP712Domain": []core.Type{
				{
					Name: "name", Type: "string"},
				{
					Name: "version", Type: "string"},
				{
					Name: "chainId", Type: "uint256"},
				{
					Name: "verifyingContract", Type: "address"},
			},
			"Permit": []core.Type{
				{
					Name: "owner", Type: "address"},
				{
					Name: "spender", Type: "address"},
				{
					Name: "value", Type: "uint256"},
				{
					Name: "nonce", Type: "uint256"},
				{
					Name: "deadline", Type: "uint256"},
			},
		},
		Domain: core.TypedDataDomain{
			Name:              name,
			Version:           version,
			ChainId:           (*math.HexOrDecimal256)(chainId),
			VerifyingContract: verifyingContract.String(),
		},
		PrimaryType: "Permit",
		Message: core.TypedDataMessage{
			"owner":    owner.String(),
			"spender":  spender.String(),
			"value":    value.String(),
			"nonce":    nonce.String(),
			"deadline": deadline.String(),
		},
	}
	signature, _, err := SignTypedData(*pk, common.MixedcaseAddress{}, data)
	if err != nil {
		logrus.Fatal(err)
	}
	v, r, s = signature[64], common.BytesToHash(signature[0:32]), common.BytesToHash(signature[32:64])
	return v, r, s
}

// todo remove commented
// func GetNodesFromContract(client *ethclient.Client, nodeListContractAddress common.Address) (nodes []wrappers.NodeRegistryNode, err error) {
// 	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
// 	if err != nil {
// 		return
// 	}
//
// 	nodes, err = nodeList.GetNodes(&bind.CallOpts{})
//
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }

// func PrintNodes(client *ethclient.Client, nodeListContractAddress common.Address) {
// 	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
// 	if err != nil {
// 		return
// 	}
//
// 	nodes, err := nodeList.GetNodes(&bind.CallOpts{})
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}
//
// 	for _, node1 := range nodes {
// 		logrus.Tracef("node ID %d node.NodeIdAddress %v", node1.NodeId, node1.NodeIdAddress)
// 	}
//
// }

// func  	GetNode(client *ethclient.Client, nodeListContractAddress common.Address, nodeBLSAddr common.Address) (node wrappers.NodeRegistryNode, err error) {
// 	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
// 	if err != nil {
// 		return
// 	}
//
// 	if exist, _ := nodeList.NodeExists(&bind.CallOpts{}, nodeBLSAddr); !exist {
// 		err = errors.New(fmt.Sprintf("NODE DOES NOT EXIST %v", nodeBLSAddr))
// 		return
// 	}
// 	node, err = nodeList.GetNode(&bind.CallOpts{}, nodeBLSAddr)
// 	if err != nil {
// 		return
// 	}
// 	return
// }

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
