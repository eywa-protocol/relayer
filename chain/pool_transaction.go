package chain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type PoolTransaction interface {
	SenderAddress() common.Address
	Nonce() uint64
	ChainId() *big.Int
	Serialize() []byte
}
