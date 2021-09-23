package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Number        int
	Signature []byte
	Leader []byte
}


// NewBlock creates and returns Block
func NewBlock(number int, transactions  []*Transaction, prevBlockHash []byte,  sig []byte, leader []byte) *Block {
	block := &Block{ time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte("wqeqwqw"),
		number,
		sig,
		leader,
	}
//TODO start block propagating protocol
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	//TODO add genesis Epoch calculation
	return NewBlock(
		1,
		[]*Transaction{coinbase},
		[]byte{},
		[]byte{},
		[]byte{})
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
