package boltdb

import (
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
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

func TestStoreBoltOrder(t *testing.T) {

	require.NoError(t, err)
	defer os.RemoveAll(tmp)
	store, err := NewBoltStore(tmp, nil)
	require.NoError(t, err)
	genesisBlock := chain.NewGenesisBlock(*header, []*chain.Transaction{coinbaseTransaction})
	b2 := chain.NewBlock(*header, 2, []*chain.Transaction{coinbaseTransaction}, []byte("sig2"))

	require.NoError(t, store.Put(genesisBlock))
	time.Sleep(10 * time.Millisecond)
	require.Equal(t, 1, store.Len())
	eb1, err := store.Last()
	require.NoError(t, err)
	t.Log(genesisBlock)
	require.Equal(t, genesisBlock, eb1)

	require.NoError(t, store.Put(b2))
	eb2, err := store.Last()
	require.NoError(t, err)
	require.Equal(t, b2, eb2)
	eb2, err = store.Last()
	require.NoError(t, err)
	require.Equal(t, b2, eb2)
	eb2, err = store.Last()
	require.NoError(t, err)
	require.Equal(t, eb2, b2)
}

func TestStoreBolt(t *testing.T) {
	tmp, err := ioutil.TempDir("", "bolttest*")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)

	var sig1 = []byte{0x01, 0x02, 0x03}
	var sig2 = []byte{0x02, 0x03, 0x04}
	var sig3 = []byte{0x03, 0x04, 0x05}

	b1 := chain.NewBlock(*header, 1, []*chain.Transaction{coinbaseTransaction}, sig1)
	b2 := chain.NewBlock(*header, 2, []*chain.Transaction{coinbaseTransaction}, sig2)
	b3 := chain.NewBlock(*header, 3, []*chain.Transaction{coinbaseTransaction}, sig3)

	store, err := NewBoltStore(tmp, nil)
	require.NoError(t, err)

	require.Equal(t, 0, store.Len())

	require.NoError(t, store.Put(b1))
	require.Equal(t, 1, store.Len())

	require.NoError(t, store.Put(b2))
	require.Equal(t, 2, store.Len())

	require.NoError(t, store.Put(b2))
	require.Equal(t, 2, store.Len())

	require.NoError(t, store.Put(b3))
	require.Equal(t, 3, store.Len())

	received, err := store.Last()
	require.NoError(t, err)
	require.Equal(t, b3, received)

	store.Close()
	store, err = NewBoltStore(tmp, nil)
	require.NoError(t, err)
	require.NoError(t, store.Put(b1))

	require.NoError(t, store.Put(b1))
	bb1, err := store.Get(1)
	require.NoError(t, err)
	require.Equal(t, b1, bb1)
	store.Close()

}

func TestStoreBoltCursor(t *testing.T) {
	var sig1 = []byte{0x02, 0x03, 0x04}
	b1 := chain.NewBlock(*header, 1, []*chain.Transaction{coinbaseTransaction}, sig1)

	var sig2 = []byte{0x02, 0x03, 0x04}
	b2 := chain.NewBlock(*header, 2, []*chain.Transaction{coinbaseTransaction}, sig2)

	store, err := NewBoltStore(tmp, nil)
	require.NoError(t, err)

	store.Put(b1)
	store.Put(b2)

	store.Cursor(func(c chain.Cursor) {
		expecteds := []*chain.Block{b1, b2}
		i := 0
		for b := c.First(); b != nil; b = c.Next() {
			require.True(t, expecteds[i].Equal(b))
			i++
		}

		unknown := c.Seek(1000)
		require.Nil(t, unknown)
	})

	store.Cursor(func(c chain.Cursor) {
		lb2 := c.Last()
		require.NotNil(t, lb2)
		require.Equal(t, b2, lb2)
	})

	unknown, err := store.Get(10000)
	require.Nil(t, unknown)
	require.Equal(t, ErrNoBlockSaved, err)

}
