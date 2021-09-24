package chain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/common"
)

// Block represents a block in the blockchain
type Block struct {
	Header       Header
	Number       uint64
	Hash         common.Hash
	Transactions []*Transaction
	Signature    []byte
	Leader       []byte
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(header Header, txs []*Transaction) *Block {
	return NewBlock(
		header,
		0,
		txs,
		[]byte("SIGNATURE"),
	)
}

// NewBlock creates and returns Block
func NewBlock(header Header, number uint64, transactions []*Transaction, sig []byte) *Block {
	block := &Block{}
	block.Header = header
	block.Number = number
	block.Transactions = transactions
	block.Signature = sig
	block.Hash = common.BytesToHash(block.HashTransactions())
	return block
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

// Marshal provides a JSON encoding of a beacon
func (b *Block) Marshal() ([]byte, error) {
	return json.Marshal(b)
}

// Unmarshal decodes a beacon from JSON
func (b *Block) Unmarshal(buff []byte) error {
	return json.Unmarshal(buff, b)
}

// Equal indicates if two beacons are equal
func (b *Block) Equal(b2 *Block) bool {
	return bytes.Equal(b.Signature, b2.Signature) &&
		b.Header.Fields.Epoch.Number == b2.Header.Fields.Epoch.Number &&
		bytes.Equal(b.Signature, b2.Signature)
}
