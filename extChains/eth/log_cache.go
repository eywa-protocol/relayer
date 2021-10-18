package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
)

type LogCache struct {
	blockNumber uint64
	data        map[uint64]map[string]struct{}
	mx          *sync.Mutex
}

func NewLogCache() *LogCache {
	return &LogCache{
		data: make(map[uint64]map[string]struct{}, 2),
		mx:   new(sync.Mutex),
	}
}

func (l *LogCache) BlockNumber() uint64 {
	l.mx.Lock()
	defer l.mx.Unlock()
	return l.blockNumber
}

func (l *LogCache) Exists(log types.Log) bool {
	l.mx.Lock()
	defer l.mx.Unlock()
	if l.blockNumber <= log.BlockNumber {
		l.blockNumber = log.BlockNumber
		l.cleanup()

		return l.addTxExists(log.BlockNumber, log.TxHash)
	} else {

		return l.addTxExists(log.BlockNumber, log.TxHash)
	}
}

func (l *LogCache) addTxExists(blockNumber uint64, txHash common.Hash) bool {
	var txExists bool
	if txMap, blockExists := l.data[blockNumber]; blockExists {
		if _, txExists = txMap[txHash.Hex()]; txExists {

			return true
		} else {
			txMap[txHash.Hex()] = struct{}{}

			return false
		}
	} else {
		txMap := make(map[string]struct{}, 1)
		txMap[txHash.Hex()] = struct{}{}
		l.data[blockNumber] = txMap

		return false
	}
}

func (l *LogCache) cleanup() {
	for u := range l.data {
		if u < l.blockNumber {
			delete(l.data, u)
		}
	}
}
