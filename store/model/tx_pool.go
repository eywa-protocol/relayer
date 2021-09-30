package model

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type PoolTransaction interface {
	SenderAddress() (common.Address, error)
	Nonce() uint64
	ChainId() *big.Int
}

// PoolTransactions is a PoolTransactions slice type for basic sorting.
type PoolTransactions []PoolTransaction

// Len returns the length of s.
func (s PoolTransactions) Len() int { return len(s) }


// PoolTxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type PoolTxByNonce PoolTransactions

func (s PoolTxByNonce) Len() int           { return len(s) }
func (s PoolTxByNonce) Less(i, j int) bool { return (s[i]).Nonce() < (s[j]).Nonce() }
func (s PoolTxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
