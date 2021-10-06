package rawdb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
)

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db DatabaseWriter, block *chain.Block) {
	// TODO: remove this hack with Tx and StakingTx structure unitification later
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
		logrus.Print("HASH", tx.Hash())
		logrus.Print("DATA", data)
			putErr = db.Put(tx.Hash(), data)
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
func ReadTransaction(db DatabaseReader, hash []byte) (chain.Transaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex := ReadTxLookupBytesEntry(db, hash)
	logrus.Print("============ > > > > " ,blockHash, blockNumber, txIndex)
	if blockHash == (common.Hash{}) {
		return chain.Transaction{}, common.Hash{}, 0, 0
	}


	bodyBytes := ReadBodyRawByNumber(db, blockNumber)
	logrus.Print("BODYBYTES", bodyBytes)
	block := chain.DeserializeBlock(bodyBytes)

	//if err := rlp.Decode(bytes.NewReader(bodyBytes), body); err != nil {
	//	logrus.Panicf(fmt.Sprintf(" - - - - > hash %v Invalid block body RLP", hash))
	//
	//}
	//body := ReadBody(db, blockHash, blockNumber)
	if block == nil {
		//utils.Logger().Error().
		//	Uint64("number", blockNumber).
		//	Str("hash", blockHash.Hex()).
		//	Uint64("index", txIndex).
		//	Msg("block Body referenced missing")
		logrus.Error(errors.New("block NIL"))
		return chain.Transaction{}, common.Hash{}, 0, 0
	}
	tx := block.Transactions[txIndex]
	if !bytes.Equal(hash, tx.Hash()) {
	//	//utils.Logger().Error().
	//	//	Uint64("number", blockNumber).
	//	//	Str("hash", blockHash.Hex()).
	//	//	Uint64("index", txIndex).
	//	//	Msg("Transaction referenced missing")
		logrus.Error(errors.New("hash not match"))
		return chain.Transaction{}, common.Hash{}, 0, 0
	}
	logrus.Print("body.Transactions", block)
	return  tx, blockHash, blockNumber, txIndex
}



// ReadTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func ReadTxLookupBytesEntry(db DatabaseReader, hash []byte) (common.Hash, uint64, uint64) {
	data, _ := db.Get(hash)
	if len(data) == 0 {
		return common.Hash{}, 0, 0
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		logrus.Printf(fmt.Sprintf("hash  Invalid transaction lookup entry ", hash))
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
		logrus.Printf(fmt.Sprintf("hash  Invalid transaction lookup entry ", hash.Hex()))
		return common.Hash{}, 0, 0
	}
	return entry.BlockHash, entry.BlockIndex, entry.Index
}

// ReadBody retrieves the block body corresponding to the hash.


func ReadBodyRawByNumber(db DatabaseReader, number uint64) rlp.RawValue {
	data, _ := db.Get([]byte(fmt.Sprint(number)))
	return data
}

//store.Get([]byte(fmt.Sprint(genesisBlock.Number)))

// ReadBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func ReadBodyRaw(db DatabaseReader, hash []byte) rlp.RawValue {
	data, _ := db.Get(hash)
	return data
}
