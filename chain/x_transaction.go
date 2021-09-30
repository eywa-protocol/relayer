package chain

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type CrossChainTransaction struct {
	chainId    *big.Int
	OriginData *wrappers.BridgeOracleRequest
	nonce      uint64
}

// NewTransaction returns new transaction,
func NewTransaction(ev *wrappers.BridgeOracleRequest, _nonce uint64, chId *big.Int) *CrossChainTransaction {
	return &CrossChainTransaction{
		OriginData: ev,
		nonce:      _nonce,
		chainId:    chId,
	}
}

func (tx *CrossChainTransaction) SenderAddress() common.Address {
	return tx.OriginData.Bridge
}
func (tx *CrossChainTransaction) Nonce() uint64 {
	return tx.nonce
}
func (tx *CrossChainTransaction) ChainId() *big.Int {
	return tx.chainId
}

// Transactions is a Transactions slice type for basic sorting.
type Transactions []*CrossChainTransaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Serialize returns a serialized Transaction
func (tx CrossChainTransaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}
