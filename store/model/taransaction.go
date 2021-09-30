package model

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Transaction struct {
    chainId    *big.Int
    OriginData *wrappers.BridgeOracleRequest
    nonce      uint64

}

// NewTransaction returns new transaction,
func NewTransaction(ev *wrappers.BridgeOracleRequest, _nonce uint64, chId *big.Int) *Transaction {
   return &Transaction{
      OriginData: ev,
      nonce: _nonce,
      chainId: chId,
   }
}

func (tx *Transaction) SenderAddress() (common.Address, error){
    return tx.OriginData.Bridge, nil
}
func (tx *Transaction) Nonce() uint64 {
    return tx.nonce
}
func (tx *Transaction) ChainId() *big.Int {
    return tx.chainId
}


// Transactions is a Transactions slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }
