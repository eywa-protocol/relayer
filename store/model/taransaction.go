package model

import (
    "github.com/ethereum/go-ethereum/common"
    "gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Transaction struct {
    OriginData *wrappers.BridgeOracleRequest
    nonce      uint64
}

// NewTransaction returns new transaction,
func NewTransaction(ev *wrappers.BridgeOracleRequest, _nonce uint64) *Transaction {
   return &Transaction{
      OriginData: ev,
      nonce: _nonce,
   }
}

func (tx *Transaction) SenderAddress() (common.Address, error){
    return tx.OriginData.Bridge, nil
}
func (tx *Transaction) Nonce() uint64 {
    return tx.nonce
}

