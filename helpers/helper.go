package helpers

import (
	"context"
	"math/big"

	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"crypto/ecdsa"
)

type OracleRequest struct {
	RequestType    string
	Bridge         common.Address
	RequestId      [32]byte
	Selector       []byte
	ReceiveSide    common.Address
	OppositeBridge common.Address
}

func FilterOracleRequestEvent(client ethclient.Client, start uint64, contractAddress common.Address) (oracleRequest OracleRequest, err error) {
	mockFilterer, err := wrappers.NewBridge(contractAddress, &client)
	if err != nil {
		return
	}

	it, err := mockFilterer.FilterOracleRequest(&bind.FilterOpts{Start: start, Context: context.Background()})
	if err != nil {
		return
	}

	for it.Next() {
		logrus.Print("OracleRequest Event", it.Event.Raw)
		if it.Event != nil {
			oracleRequest = OracleRequest{
				RequestType:    it.Event.RequestType,
				Bridge:         it.Event.Bridge,
				RequestId:      it.Event.RequestId,
				Selector:       it.Event.Selector,
				ReceiveSide:    it.Event.ReceiveSide,
				OppositeBridge: it.Event.OppositeBridge,
			}
		}
	}
	return
}

/**
* Catch event from first side
 */
func ListenOracleRequest(
	clientNetwork_1 *ethclient.Client,
	clientNetwork_2 *ethclient.Client,
	proxyNetwork_1 common.Address,
	proxyNetwork_2 common.Address) (oracleRequest OracleRequest, err error) {

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

				privateKey, err := crypto.HexToECDSA("95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc")
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

				oracleRequest = OracleRequest{
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

//TODO: sinchroniize between ListenOracleRequest
func ListenReceiveRequest(clientNetwork *ethclient.Client, proxyNetwork common.Address) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork, clientNetwork)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeReceiveRequest)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchReceiveRequest(opt, channel)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case err := <-sub.Err():
				logrus.Println("ReceiveRequest error:", err)
			case event := <-channel:
				logrus.Printf("ReceiveRequest: ", event.ReqId, event.ReceiveSide, event.Tx)

				/** TODO:
				Is transaction true, otherwise repeate to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}
