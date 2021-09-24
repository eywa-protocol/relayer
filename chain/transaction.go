package chain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strings"
)

type TxType int

const (
	ChangeEpochTx TxType = iota // ChangeEpoch 0
	EventTx       TxType = 1    // Events     []*wrappers.BridgeOracleRequest 1
	UptimeTx      TxType = 2    // UptimeList uptime.UpList 2
)

// Transaction represents a EYWA transaction = EYWA bridge cross-chain action abstraction
type Transaction struct {
	ID      []byte
	Type    TxType
	Payload []byte
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return string(tx.ID) == "0"
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
	txCopy.ID = []byte{}

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
	tx := Transaction{epochId, ChangeEpochTx, nil}
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
