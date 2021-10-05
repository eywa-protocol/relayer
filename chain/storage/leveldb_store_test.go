package storage

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain/rawdb"
	"io/ioutil"
	"os"
	"testing"
	"time"
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
	require.NoError(t, store.Put([]byte(fmt.Sprint(genesisBlock.Number)), genesisBlock.Serialize()))
	time.Sleep(10 * time.Millisecond)
	rawdb.WriteTxLookupEntries(store, genesisBlock)
	getGenesisBlock, err := store.Get([]byte(fmt.Sprint(genesisBlock.Number)))
	require.NoError(t, err)
	t.Log(genesisBlock)
	readBlock := chain.DeserializeBlock(getGenesisBlock)
	t.Log(readBlock)

	require.Equal(t, genesisBlock, readBlock)

	txs := readBlock.Transactions

	for _, tx := range txs {
		//readtx, hash, _, _ := rawdb.ReadTransaction(store, tx.Hash())
		//t.Log(readtx, hash)
		t.Log(tx.OriginData)
	}

}
