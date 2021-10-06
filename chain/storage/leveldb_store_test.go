package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain/rawdb"
)

var (
	tmp, err            = ioutil.TempDir("", "test*")
	genesisEpoch        = chain.CreateGenesisEpoch()
	header              = chain.NewHeader(*genesisEpoch)
	coinbaseTransaction = chain.NewCoinbaseTX(genesisEpoch.Serialize())
)

func newLevelDBForTesting(t *testing.T) Store {
	ldbDir := t.TempDir()
	dbConfig := DBConfiguration{
		Type: "leveldb",
		LevelDBOptions: LevelDBOptions{
			DataDirectoryPath: ldbDir,
		},
	}
	newLevelStore, err := NewLevelDBStore(dbConfig.LevelDBOptions)
	require.Nil(t, err, "NewLevelDBStore error")
	return newLevelStore
}

func TestPutAndGetGenesisBlock(t *testing.T) {

	require.NotNil(t, coinbaseTransaction)
	require.NoError(t, err)

	tmp, err := ioutil.TempDir("", "leveldb*")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)

	store := newLevelDBForTesting(t)
	require.NoError(t, err)
	t.Log(genesisEpoch.Serialize())
	t.Log(genesisEpoch)
	t.Log(header)

	t.Log(coinbaseTransaction.OriginData)
	genesisBlock := chain.NewGenesisBlock(*header, []chain.Transaction{*coinbaseTransaction})
	t.Log(common.BytesToHash(genesisBlock.HashTransactions()))
	rawdb.WriteBlock(store, genesisBlock)

	//require.NoError(t, store.Put([]byte(fmt.Sprint(genesisBlock.Number)), genesisBlock.Serialize()))
	rawdb.WriteTxLookupEntries(store, genesisBlock)
	//getGenesisBlock, err := store.Get([]byte(fmt.Sprint(genesisBlock.Number)))

	getGenesisBlock := rawdb.ReadBody(store, genesisBlock.Hash, 1)

	require.NoError(t, err)
	t.Log(getGenesisBlock)
	//readBlock := chain.(getGenesisBlock)
	//t.Log(readBlock)

	require.Equal(t, genesisBlock, &getGenesisBlock)

	//txs := readBlock.Transactions

	//for _, tx := range txs {
	//	t.Log(tx.Hash())
	//	readtx, hash, _, _ := rawdb.ReadTransaction(store, tx.Hash())
	//	t.Log(readtx, hash)
	//	t.Log(tx.OriginData)
	//	require.Equal(t, *coinbaseTransaction, readtx)
	//}
	readtx, _, _, _ := rawdb.ReadTransaction(store, coinbaseTransaction.Hash())
	require.Equal(t, getGenesisBlock.Hash,  common.BytesToHash(coinbaseTransaction.Hash()))
	require.Equal(t, *coinbaseTransaction, readtx)

}
