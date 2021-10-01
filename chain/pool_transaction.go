package chain

import (
	"github.com/ethereum/go-ethereum/common"
)

type PoolTransaction interface {
	SenderAddress() common.Address
	Nonce() uint64
	ChainId() uint64
	Serialize() []byte
}
