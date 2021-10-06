package storage

import (
	"io/ioutil"
	"os"
	"testing"

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

func TestPutAndGetGenesisBlockFunction(t *testing.T) {

	require.NotNil(t, coinbaseTransaction)
	require.NoError(t, err)

	tmp, err := ioutil.TempDir("", "leveldb*")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)

	store := newLevelDBForTesting(t)
	require.NoError(t, err)

	//t.Log(genesisEpoch.Serialize())
	//t.Log(genesisEpoch)
	//t.Log(header)
	//t.Log(coinbaseTransaction.OriginData)

	genesisBlock := chain.NewGenesisBlock(*header, []chain.Transaction{*coinbaseTransaction})
	//t.Log(common.BytesToHash(genesisBlock.HashTransactions()))
	rawdb.WriteBlock(store, genesisBlock.Hash, genesisBlock.Number, *genesisBlock)

	rawdb.WriteTxLookupEntries(store, genesisBlock)

	readBlock := rawdb.ReadBlock(store, genesisBlock.Hash, genesisBlock.Number)

	require.NotNil(t, readBlock)
	require.NoError(t, err)
	//t.Log(readBlock)

	require.Equal(t, genesisBlock, readBlock)

	readtx, blockHash, _, _ := rawdb.ReadTransaction(store, coinbaseTransaction.Hash())
	//t.Log(readtx)
	//t.Log(blockHash)
	require.Equal(t, readBlock.Hash, blockHash)
	require.Equal(t, genesisBlock.Hash, blockHash)
	require.Equal(t, coinbaseTransaction, readtx)

}
