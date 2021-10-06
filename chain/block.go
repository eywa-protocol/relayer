package chain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Block represents a block in the blockchain
type Block struct {
	Number       uint64
	Hash         common.Hash
	Transactions []Transaction
	Signature    []byte
	Leader       []byte
	ChanId       big.Int
	Header       *Header
}

//type Block struct {
//	header Header
//	bm      BaseModel
//	tx        []Transaction
//	chanId int
//}

type BaseModel struct {
	id int
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(header Header, txs []Transaction) *Block {
	return NewBlock(
		header,
		1,
		txs,
		[]byte("SIGNATURE"),
	)
}

// NewBlock creates and returns Block
func NewBlock(header Header, number uint64, transactions []Transaction, sig []byte) *Block {
	block := &Block{}
	block.Header = &header
	block.Number = number
	block.Transactions = transactions
	block.Signature = sig
	block.Hash = common.BytesToHash(block.HashTransactions())
	block.Leader = []byte{}
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
		logrus.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		logrus.Panic(err)
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
		b.Header.Fields.Epoch.Number.Cmp(&b2.Header.Fields.Epoch.Number) == 0 &&
		bytes.Equal(b.Signature, b2.Signature)
}

func CopyHeader(header Header) *Header {
	return &header
}

// Body is a simple (mutable, non-safe) data container for storing and moving
// a block's data contents (transactions and uncles) together.
type Body struct {
	BodyInterface
}

// BodyInterface is a simple accessor interface for block body.
type BodyInterface interface {
	// Transactions returns a deep copy the list of transactions in this block.
	Transactions() []*Transaction

	// TransactionAt returns the transaction at the given index in this block.
	// It returns nil if index is out of bounds.
	TransactionAt(index int) *Transaction

	// SetTransactions sets the list of transactions with a deep copy of the
	// given list.
	SetTransactions(newTransactions []*Transaction)

	// Uncles returns a deep copy of the list of uncle headers of this block.
	Uncles() []*Header

	// SetUncles sets the list of uncle headers with a deep copy of the given
	// list.
	SetUncles(newUncle []*Header)
}
