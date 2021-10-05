package rawdb

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
func ReadTransaction(db DatabaseReader, hash []byte) (*chain.Transaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex := ReadTxLookupEntry(db, common.BytesToHash(hash))
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	body := ReadBody(db, blockHash, blockNumber)
	if body == nil {
		//utils.Logger().Error().
		//	Uint64("number", blockNumber).
		//	Str("hash", blockHash.Hex()).
		//	Uint64("index", txIndex).
		//	Msg("block Body referenced missing")
		return nil, common.Hash{}, 0, 0
	}
	//tx := body.TransactionAt(int(txIndex))
	//if tx == nil || !bytes.Equal(hash.Bytes(), tx.Hash().Bytes()) {
	//	//utils.Logger().Error().
	//	//	Uint64("number", blockNumber).
	//	//	Str("hash", blockHash.Hex()).
	//	//	Uint64("index", txIndex).
	//	//	Msg("Transaction referenced missing")
	//	return nil, common.Hash{}, 0, 0
	//}
	return  body.Transactions[], blockHash, blockNumber, txIndex
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
func ReadBody(db DatabaseReader, hash common.Hash, number uint64) *types.Body {
	data := ReadBodyRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	body := new(types.Body)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		logrus.Errorf(fmt.Sprintf("hash %v Invalid block body RLP", hash.Hex()))
		return nil
	}
	return body
}

// ReadBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func ReadBodyRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(blockBodyKey(number, hash))
	return data
}
