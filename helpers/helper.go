package helpers

import (
	"context"
	"strings"
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	ethereum "github.com/ethereum/go-ethereum"
	"math/big"
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

func ListenOracleRequest(client *ethclient.Client, contractAddress common.Address) (oracleRequest OracleRequest, err error) {
	bridgeFilterer, err := wrappers.NewBridge(contractAddress, client)
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

func WorkerEvent(client *ethclient.Client, _contractAddress string) {

	contractAddress := common.HexToAddress(_contractAddress)
    query 			:= ethereum.FilterQuery{
					        FromBlock: big.NewInt(20),
					        ToBlock:   nil,
					        Addresses: []common.Address{
					            contractAddress,
					        },
					    }

    logs, err := client.FilterLogs(context.Background(), query)
    	if err != nil { logrus.Fatal(err) }		
	contractAbi, err := abi.JSON(strings.NewReader(string(wrappers.BridgeABI)))
    	if err != nil {logrus.Fatal(err)}			    
	eventSignature := []byte("OracleRequest(string,address,bytes32,bytes,address,address)")
    	eventOracleRequest := crypto.Keccak256Hash(eventSignature)
	
	for _, vLog := range logs {
        logrus.Info("Log Block Number: ", vLog.BlockNumber)
		logrus.Info("Log Top: ",  vLog.Topics)

        switch vLog.Topics[0].Hex() {

        	case eventOracleRequest.Hex(): 
        		  logrus.Info("EventOracleRequest triggered")
        		  
				  res, err := contractAbi.Unpack("OracleRequest", vLog.Data)
				  if err != nil {logrus.Fatal(err)}

  				  var oracleRequest = OracleRequest{
						RequestType:    res[0].(string),
						Bridge:         res[1].(common.Address),
						RequestId:      res[2].([32]byte),
						Selector:       res[3].([]byte),
						ReceiveSide:    res[4].(common.Address),
						OppositeBridge: res[5].(common.Address),
					}
				  
				  logrus.Info("DEBUG", oracleRequest)
				  
        }
	}

    _=contractAbi
    _=contractAddress
	
	//logrus.Info("DEBUG", eventOracleRequest.Hex())
	//logrus.Info("DEBUG", logs)

}
