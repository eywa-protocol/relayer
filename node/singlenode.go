package node

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"strings"

	wrappers "github.com/digiu-ai/eth-contracts/wrappers"
	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type PendingRequest struct {
	Tx      types.Transaction
	Event   helpers.OracleRequest
	Reciept types.Receipt
	From    common.Address
}

func NewSingleNode(path string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}
	/** if ENV empty then use config */
	if os.Getenv("ECDSA_KEY_1") != "" || os.Getenv("ECDSA_KEY_2") != "" {
		config.Config.ECDSA_KEY_1 = strings.Split(os.Getenv("ECDSA_KEY_1"), ",")[0]
		config.Config.ECDSA_KEY_2 = strings.Split(os.Getenv("ECDSA_KEY_2"), ",")[0]
	}
	/**  */
	if os.Getenv("NETWORK_RPC_1") != "" || os.Getenv("NETWORK_RPC_2") != "" {
		config.Config.NETWORK_RPC_1 = os.Getenv("NETWORK_RPC_1")
		config.Config.NETWORK_RPC_2 = os.Getenv("NETWORK_RPC_2")
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx: ctx,
	}

	server := n.NewBridge()
	n.Server = *server

	n.EthClient_1, n.EthClient_2, _ = getEthClients()

	pendingTxFromNetwork1 := make(chan *PendingRequest)
	pendingTxFromNetwork2 := make(chan *PendingRequest)

	var pendingListWaitRecieptFromNetwork1 = make(map[common.Hash]PendingRequest)
	var pendingListWaitRecieptFromNetwork2 = make(map[common.Hash]PendingRequest)

	/**
	* subscribe on events newtwork1
	 */
	_, err = subscNodeOracleRequest(
		n.EthClient_1,
		n.EthClient_2,
		common.HexToAddress(config.Config.PROXY_NETWORK1),
		common.HexToAddress(config.Config.PROXY_NETWORK2),
		config.Config.ECDSA_KEY_2,
		pendingTxFromNetwork2)
	if err != nil {
		panic(err)
	}
	/** Save num tx in map  */
	go func() {
		for {
			_resolvedTx := <-pendingTxFromNetwork2
			pendingListWaitRecieptFromNetwork2[_resolvedTx.Tx.Hash()] = *_resolvedTx
			logrus.Tracef("Tx: %s was added in pending list network2, receipt status: %v, sender: %s", _resolvedTx.Tx.Hash(), _resolvedTx.Reciept.Status, _resolvedTx.From)
			logrus.Tracef("The count of pending network2 store tx is: %d ", len(pendingListWaitRecieptFromNetwork2))
		}
	}()

	/**
	* subscribe on events newtwork2
	 */
	_, err = subscNodeOracleRequest(
		n.EthClient_2,
		n.EthClient_1,
		common.HexToAddress(config.Config.PROXY_NETWORK2),
		common.HexToAddress(config.Config.PROXY_NETWORK1),
		config.Config.ECDSA_KEY_1,
		pendingTxFromNetwork1)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			__resolvedTx := <-pendingTxFromNetwork1
			pendingListWaitRecieptFromNetwork1[__resolvedTx.Tx.Hash()] = *__resolvedTx
			logrus.Tracef("Tx: %s was added in pending list network1, receipt status: %v, sender %s", __resolvedTx.Tx.Hash(), __resolvedTx.Reciept.Status, __resolvedTx.From)
			logrus.Tracef("The count of pending network2 store tx is: %d ", len(pendingListWaitRecieptFromNetwork1))
		}
	}()

	port := config.Config.PORT_1
	n.Server.Start(port)
	run(n.Host, cancel)

	return
}

/**
* Catch event from first side
 */
func subscNodeOracleRequest(
	clientNetwork_1 *ethclient.Client,
	clientNetwork_2 *ethclient.Client,
	proxyNetwork_1 common.Address,
	proxyNetwork_2 common.Address,
	sender string,
	pendingTx chan *PendingRequest) (oracleRequest helpers.OracleRequest, err error) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork_1, clientNetwork_1)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeOracleRequest)
	opt := &bind.WatchOpts{}
	sub, err := bridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		logrus.Errorf("Error %v", err.Error())
		return
	}
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Errorf("OracleRequest error: %v", err)
			case event := <-channel:
				logrus.Trace("Catching OracleRequest event")

				privateKey, err := crypto.HexToECDSA(sender[2:])
				if err != nil {
					logrus.Fatal(err)
				}
				publicKey := privateKey.Public()
				publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
				if !ok {
					logrus.Fatal("error casting public key to ECDSA")
				}
				fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
				nonce, err := clientNetwork_2.PendingNonceAt(context.Background(), fromAddress)
				if err != nil {
					logrus.Fatal(err)
				}
				gasPrice, err := clientNetwork_2.SuggestGasPrice(context.Background())
				if err != nil {
					logrus.Fatal(err)
				}
				auth := bind.NewKeyedTransactor(privateKey)
				auth.Nonce = big.NewInt(int64(nonce))
				auth.Value = big.NewInt(0)     // in wei
				auth.GasLimit = uint64(300000) // in units
				//NOTE!!! TODO: should be adjusted for optimal gas cunsumption
				auth.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))

				instance, err := wrappers.NewBridge(proxyNetwork_2, clientNetwork_2)
				if err != nil {
					logrus.Fatal(err)
				}

				/**   TODO: apporove that tx was real  */
				/**   TODO: apporove that tx was real  */
				/**   TODO: apporove that tx was real  */
				/**   TODO: apporove that tx was real  */

				oracleRequest = helpers.OracleRequest{
					RequestType:    event.RequestType,
					Bridge:         event.Bridge,
					RequestId:      event.RequestId,
					Selector:       event.Selector,
					ReceiveSide:    event.ReceiveSide,
					OppositeBridge: event.OppositeBridge,
				}
				/** Invoke bridge on another side */
				tx, err := instance.ReceiveRequestV2(auth, "", nil, oracleRequest.Selector, [32]byte{}, oracleRequest.ReceiveSide)
				if err != nil {
					logrus.Error(err)
				}

				receipt, err := helpers.WaitTransaction(clientNetwork_2, tx)
				if err != nil {
					logrus.Warn("TX was refused, e.g. tx.status == fasle, ", tx.Hash())
				}

				pendingTx <- &PendingRequest{
					Tx:      *tx,
					Event:   oracleRequest,
					Reciept: *receipt,
					From:    fromAddress,
				}
			}
		}
	}()
	return
}
