package model

import "github.com/ethereum/go-ethereum/common"

type PoolTransaction interface {
	SenderAddress() (common.Address, error)
	Nonce() uint64
}

// PoolTransactions is a PoolTransactions slice type for basic sorting.
type PoolTransactions []PoolTransaction

// Len returns the length of s.
func (s PoolTransactions) Len() int { return len(s) }
