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
					RequestType:    event.RequestType,
					Bridge:         event.Bridge,
					RequestId:      event.RequestId,
					Selector:       event.Selector,
					ReceiveSide:    event.ReceiveSide,
					OppositeBridge: event.OppositeBridge,
				}
			}
		}
	}()
	return
}
