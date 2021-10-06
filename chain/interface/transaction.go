package _interface

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Transaction interface {
	SenderAddress() common.Address
	Nonce() uint64
	ChainId() *big.Int
	Serialize() []byte
}
