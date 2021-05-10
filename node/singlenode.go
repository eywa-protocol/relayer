package node

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"

	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

func NewSingleNode(path string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}
	/** parametrs does't exist in consfig */
	if config.Config.ECDSA_KEY_1 == "" {
		config.Config.ECDSA_KEY_1 = os.Getenv("ECDSA_KEY_1")
		config.Config.ECDSA_KEY_2 = os.Getenv("ECDSA_KEY_2")
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx: ctx,
	}

	server := n.NewBridge()
	n.Server = *server

	n.EthClient_1, n.EthClient_2, _ = getEthClients()
	/**
	* subscribe on events newtwork1
	 */
	_, err = subscNodeOracleRequest(
		n.EthClient_1,
		n.EthClient_2,
		common.HexToAddress(config.Config.PROXY_NETWORK1),
		common.HexToAddress(config.Config.PROXY_NETWORK2),
		config.Config.ECDSA_KEY_2)
	if err != nil {
		panic(err)
	}

	/**
	* subscribe on events newtwork2
	 */
	_, err = subscNodeOracleRequest(
		n.EthClient_2,
		n.EthClient_1,
		common.HexToAddress(config.Config.PROXY_NETWORK2),
		common.HexToAddress(config.Config.PROXY_NETWORK1),
		config.Config.ECDSA_KEY_1)
	if err != nil {
		panic(err)
	}

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
	sender string) (oracleRequest helpers.OracleRequest, err error) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork_1, clientNetwork_1)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeOracleRequest)
	opt := &bind.WatchOpts{}
	sub, err := bridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Println("OracleRequest error:", err)
			case event := <-channel:
				logrus.Printf("OracleRequest id: %d type: %d\n", event.RequestId, event.RequestType)

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
				auth.GasPrice = gasPrice

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
					logrus.Fatal(err)
				}

				logrus.Printf("tx in first chain has been triggered :  ", tx.Hash())

			}
		}
	}()
	return
}
