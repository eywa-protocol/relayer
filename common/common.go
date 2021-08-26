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
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core"
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

func RegisterNode(client *ethclient.Client, from, wallet *ecdsa.PrivateKey, nodeRegistryAddress common.Address, peerId peer.ID, blsPubkey string) (id *big.Int, relayerPool *common.Address, err error) {
	logrus.Infof("Adding Node %s it's NodeidAddress %x", peerId, common.BytesToAddress([]byte(peerId.String())))
	fromAddress := AddressFromSecp256k1PrivKey(from)
	walletAddress := AddressFromSecp256k1PrivKey(wallet)
	peerIdAsAddress := common.BytesToAddress([]byte(peerId))

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("get chain id error: %w", err)
	}
	nodeRegistry, err := wrappers.NewNodeRegistry(nodeRegistryAddress, client)
	res, err := nodeRegistry.NodeExists(&bind.CallOpts{}, peerIdAsAddress)
	if err != nil {
		err = fmt.Errorf("node not exists nodeRegistryAddress: %s, client.Id: %s, error: %w",
			nodeRegistryAddress.String(), chainId.String(), err)
	}
	if res == true {
		logrus.Infof("Node %x allready exists", peerId)
		return
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := make(chan bool)
	go func() {
		time.Sleep(15 * time.Second)
		timeout <- true
	}()
	for {
		select {
		case <-ticker.C:
			eywaAddress, _ := nodeRegistry.EYWA(&bind.CallOpts{})
			eywa, err := wrappers.NewERC20Permit(eywaAddress, client)
			if err != nil {
				return nil, nil, fmt.Errorf("EYWA contract: %w", err)
			}
			fromNonce, _ := eywa.Nonces(&bind.CallOpts{}, fromAddress)
			value, _ := eywa.BalanceOf(&bind.CallOpts{}, fromAddress)

			deadline := big.NewInt(time.Now().Unix() + 100)
			const EywaPermitName = "EYWA"
			const EywaPermitVersion = "1"
			v, r, s := signErc20Permit(from, EywaPermitName, EywaPermitVersion, chainId,
				eywaAddress, fromAddress, nodeRegistryAddress, value, fromNonce, deadline)

			node := wrappers.NodeRegistryNode{
				Owner:                 fromAddress,
				NodeWallet:            walletAddress,
				NodeIdAddress:         peerIdAsAddress,
				Vault:                 fromAddress, // fixme
				Pool:                  common.Address{},
				BlsPubKey:             blsPubkey,
				NodeId:                big.NewInt(0),
				Version:               big.NewInt(0),
				RelayerFeeNumerator:   big.NewInt(100),  // fixme
				EmissionRateNumerator: big.NewInt(4000), // fixme
				Status:                0}                // online

			txOpts1 := CustomAuth(client, from)
			tx, err := nodeRegistry.CreateRelayer(txOpts1, node, deadline, v, r, s)
			if err != nil {
				logrus.Errorf("CreateRelayer chainId %d ERROR: %v", chainId, err)
				if strings.Contains(err.Error(), "allready exists") || strings.Contains(err.Error(), "gas required exceeds allowance") {
					return nil, nil, err
				}
				break
			}
			recept, err := helpers.WaitTransaction(client, tx)
			if err != nil {
				return nil, nil, fmt.Errorf("WaitTransaction: %w", err)
			}
			logrus.Tracef("recept.Status ", recept.Status)

			blockNum := recept.BlockNumber.Uint64()
			filterer, err := wrappers.NewNodeRegistryFilterer(nodeRegistryAddress, client)
			if err != nil {
				return nil, nil, fmt.Errorf("NodeRegistry filterer: %w", err)
			}
			it, _ := filterer.FilterCreatedRelayer(&bind.FilterOpts{Start: blockNum, End: &blockNum},
				[]common.Address{node.NodeIdAddress}, []*big.Int{}, []common.Address{})
			defer it.Close()
			for it.Next() {
				return it.Event.NodeId, &it.Event.RelayerPool, nil
			}
			return nil, nil, nil

		case <-timeout:
			return nil, nil, errors.New("Timeout")
		}
	}
}

func signErc20Permit(pk *ecdsa.PrivateKey, name, version string, chainId *big.Int, verifyingContract, owner, spender common.Address, value, nonce, deadline *big.Int) (v uint8, r [32]byte, s [32]byte) {
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

func GetNodesFromContract(client *ethclient.Client, nodeListContractAddress common.Address) (nodes []wrappers.NodeRegistryNode, err error) {
	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
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
	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
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

func GetNode(client *ethclient.Client, nodeListContractAddress common.Address, nodeBLSAddr common.Address) (node wrappers.NodeRegistryNode, err error) {
	nodeList, err := wrappers.NewNodeRegistry(nodeListContractAddress, client)
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
