package rawdb

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
)

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db DatabaseWriter, block *chain.Block) {
	f := func(i int, tx *chain.Transaction) {
		entry := TxLookupEntry{
			BlockHash:  block.Hash,
			BlockIndex: block.Number,
			Index:      uint64(i),
		}
		data, err := rlp.EncodeToBytes(entry)
		if err != nil {
			logrus.Error(err)
		}
		var putErr error
		putErr = db.Put(txLookupKey(tx.Hash()), data)
		if putErr != nil {
			logrus.Errorf("Failed to store transaction lookup entry: %v", putErr)
		}
	}
	for i, tx := range block.Transactions {
		f(i, &tx)
	}
}

// ReadTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.

func ReadTransaction(db DatabaseReader, hash common.Hash) (*chain.Transaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex := ReadTxLookupEntry(db, hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	body := ReadBlock(db, blockHash, blockNumber)
	if body == nil {
		logrus.Fatal("block Body referenced missing")
		return nil, common.Hash{}, 0, 0
	}
	tx := body.Transactions[txIndex]
	return &tx, blockHash, blockNumber, txIndex
}

// ReadTxLookupBytesEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func ReadTxLookupBytesEntry(db DatabaseReader, hash []byte) (common.Hash, uint64, uint64) {
	data, _ := db.Get(hash)
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		logrus.Printf("hash  Invalid transaction lookup entry %x", hash)
		return common.Hash{}, 0, 0
	}
	return entry.BlockHash, entry.BlockIndex, entry.Index
}

// ReadTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func ReadTxLookupEntry(db DatabaseReader, hash common.Hash) (common.Hash, uint64, uint64) {
	data, _ := db.Get(txLookupKey(hash))
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		logrus.Printf("hash  Invalid transaction lookup entry %s", hash.Hex())
		return common.Hash{}, 0, 0
	}
	return entry.BlockHash, entry.BlockIndex, entry.Index
}

// ReadBody retrieves the block body corresponding to the hash.

func ReadBodyRawByNumber(db DatabaseReader, number uint64) rlp.RawValue {
	data, _ := db.Get([]byte(fmt.Sprint(number)))
	return data
}

// store.Get([]byte(fmt.Sprint(genesisBlock.Number)))

// ReadBodyRaw retrieves the block body (transactions and uncles) in RLP encoding.
func ReadBodyRaw(db DatabaseReader, hash []byte) rlp.RawValue {
	data, _ := db.Get(hash)
	return data
}
