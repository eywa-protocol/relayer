package helpers

import (
	"context"
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type OracleRequest struct {
	requestType    string
	bridge         common.Address
	requestId      [32]byte
	selector       []byte
	receiveSide    common.Address
	oppositeBridge common.Address
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
				requestType:    it.Event.RequestType,
				bridge:         it.Event.Bridge,
				requestId:      it.Event.RequestId,
				selector:       it.Event.Selector,
				receiveSide:    it.Event.ReceiveSide,
				oppositeBridge: it.Event.OppositeBridge,
			}
		}
	}
	return
}

func ListenOracleRequest(client ethclient.Client, contractAddress common.Address) (oracleRequest OracleRequest, err error) {
	bridgeFilterer, err := wrappers.NewBridge(contractAddress, &client)
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
				oracleRequest = OracleRequest{
					requestType:    event.RequestType,
					bridge:         event.Bridge,
					requestId:      event.RequestId,
					selector:       event.Selector,
					receiveSide:    event.ReceiveSide,
					oppositeBridge: event.OppositeBridge,
				}
			}
		}
	}()
	return
}
