package eth

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type newEpochWatcher struct {
	contractAbi     abi.ABI
	contractAddress common.Address
	eventHandler    NewEpochHandler
}

type NewEpochHandler func(event *wrappers.BridgeNewEpoch, srcChainId *big.Int)

func NewEpochWatcher(address common.Address, eventHandler NewEpochHandler) (*newEpochWatcher, error) {
	if eventHandler == nil {

		return nil, ErrHandlerUndefined
	} else if contractAbi, err := abi.JSON(strings.NewReader(wrappers.BridgeABI)); err != nil {

		return nil, err
	} else {

		return &newEpochWatcher{
			contractAbi:     contractAbi,
			contractAddress: address,
			eventHandler:    eventHandler,
		}, nil
	}

}

func (o newEpochWatcher) Abi() abi.ABI {
	return o.contractAbi
}

func (o newEpochWatcher) Name() string {

	return "NewEpoch"
}

func (o newEpochWatcher) Address() common.Address {
	return o.contractAddress
}

func (o newEpochWatcher) Query() [][]interface{} {
	return nil
}

func (o newEpochWatcher) NewEventPointer() interface{} {
	return &wrappers.BridgeNewEpoch{}
}

func (o newEpochWatcher) SetEventRaw(eventPointer interface{}, log types.Log) {
	eventPointer.(*wrappers.BridgeNewEpoch).Raw = log
}

func (o newEpochWatcher) OnEvent(eventPointer interface{}, srcChainId *big.Int) {
	if req, ok := eventPointer.(*wrappers.BridgeNewEpoch); !ok {
		logrus.Error(ErrUnsupportedEvent)
	} else {
		logrus.Infof("bridge NewEpoch event received tx: %s", req.Raw.TxHash.Hex())
		o.eventHandler(req, srcChainId)
	}
}
