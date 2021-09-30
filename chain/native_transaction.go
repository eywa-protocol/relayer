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
)

//type TxType int
//
//const (
//	ChangeEpochTx TxType = iota // ChangeEpoch 0
//	EventTx       TxType = 1    // Events     []*wrappers.BridgeOracleRequest 1
//	UptimeTx      TxType = 2    // UptimeList uptime.UpList 2
//)

// Transaction represents a EYWA transaction = EYWA bridge cross-chain action abstraction
type NativeTransaction struct {
	ID      []byte
	Type    TxType
	Payload []byte
}

func (tx *NativeTransaction) SenderAddress() common.Address {
	return common.HexToAddress("0x0")
}
func (tx *NativeTransaction) Nonce() uint64 {
	return 0
}
func (tx *NativeTransaction) ChainId() *big.Int {
	return big.NewInt(0)
}

// IsCoinbase checks whether the transaction is coinbase
func (tx NativeTransaction) IsCoinbase() bool {
	return string(tx.ID) == "0"
}

// Serialize returns a serialized Transaction
func (tx NativeTransaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *NativeTransaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *NativeTransaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]NativeTransaction) {
	if tx.IsCoinbase() {
		return
	}
	//TODO Sign tx with node BLS key

}

// String returns a human-readable representation of a transaction
func (tx NativeTransaction) String() string {
	var lines []string
	//TDOD Implement tx String method
	return strings.Join(lines, "\n")
}

// Verify verifies signatures of Transaction inputs
func (tx *NativeTransaction) Verify(txs map[string]NativeTransaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//TODO Verify transactions signed by nodes

	return true
}

//
//// NewCoinbaseTX creates a new coinbase transaction
//func NewCoinbaseTX(epochId []byte) *NativeTransaction {
//	tx := NativeTransaction{epochId, ChangeEpochTx, nil}
//	return &tx
//}
//
//// DeserializeTransaction deserializes a transaction
//func DeserializeTransaction(data []byte) NativeTransaction {
//	var transaction NativeTransaction
//
//	decoder := gob.NewDecoder(bytes.NewReader(data))
//	err := decoder.Decode(&transaction)
//	if err != nil {
//		log.Panic(err)
//	}
//
//	return transaction
//}
