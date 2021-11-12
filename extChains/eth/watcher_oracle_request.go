package eth

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
	"strings"
)

type oracleRequestWatcher struct {
	contractAbi     abi.ABI
	contractAddress common.Address
	eventHandler    OracleRequestHandler
}

type OracleRequestHandler func(event *wrappers.BridgeOracleRequest, srcChainId *big.Int)

func NewOracleRequestWatcher(address common.Address, eventHandler OracleRequestHandler) (*oracleRequestWatcher, error) {
	if eventHandler == nil {

		return nil, ErrHandlerUndefined
	} else if contractAbi, err := abi.JSON(strings.NewReader(wrappers.BridgeABI)); err != nil {

		return nil, err
	} else {

		return &oracleRequestWatcher{
			contractAbi:     contractAbi,
			contractAddress: address,
			eventHandler:    eventHandler,
		}, nil
	}

}

func (o oracleRequestWatcher) Abi() abi.ABI {
	return o.contractAbi
}

func (o oracleRequestWatcher) Name() string {

	return "OracleRequest"
}

func (o oracleRequestWatcher) Address() common.Address {
	return o.contractAddress
}

func (o oracleRequestWatcher) Query() [][]interface{} {
	return nil
}

func (o oracleRequestWatcher) NewEventPointer() interface{} {
	return &wrappers.BridgeOracleRequest{}
}

func (o oracleRequestWatcher) SetEventRaw(eventPointer interface{}, log types.Log) {
	eventPointer.(*wrappers.BridgeOracleRequest).Raw = log
}

func (o oracleRequestWatcher) OnEvent(eventPointer interface{}, srcChainId *big.Int) {
	if req, ok := eventPointer.(*wrappers.BridgeOracleRequest); !ok {
		logrus.Error(ErrUnsupportedEvent)
	} else {
		logrus.Infof("bridge oracle request received tx: %s", req.Raw.TxHash.Hex())
		o.eventHandler(req, srcChainId)
	}
}
