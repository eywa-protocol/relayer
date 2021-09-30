package chain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type TxType int

const (
	ChangeEpochTx TxType = iota // ChangeEpoch 0
	EventTx       TxType = 1    // Events     []*wrappers.BridgeOracleRequest 1
	UptimeTx      TxType = 2    // UptimeList uptime.UpList 2
)

// Transaction represents a EYWA transaction = EYWA bridge cross-chain action abstraction
type Transaction struct {
	chainId    *big.Int
	OriginData *wrappers.BridgeOracleRequest
	nonce      uint64
}

// NewTransaction returns new transaction,
func NewTransaction(ev *wrappers.BridgeOracleRequest, _nonce uint64, chId *big.Int) *Transaction {
	return &Transaction{
		OriginData: ev,
		nonce:      _nonce,
		chainId:    chId,
	}
}

func (tx *Transaction) SenderAddress() common.Address {
	return tx.OriginData.Bridge
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

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return tx.nonce == uint64(0) && tx.chainId.Cmp(big.NewInt(0)) == 0
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	//TODO Sign tx with node BLS key

}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string
	//TDOD Implement tx String method
	return strings.Join(lines, "\n")
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(txs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//TODO Verify transactions signed by nodes

	return true
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(epochId []byte) *Transaction {
	tx := Transaction{
		big.NewInt(0),
		&wrappers.BridgeOracleRequest{},
		uint64(0),
	}
	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
