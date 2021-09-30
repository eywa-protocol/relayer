package store

import (
	"container/heap"
	_ "math/big"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/model"
)


// nonceHeap is a heap.Interface implementation over 64bit unsigned integers for
// retrieving sorted transactions from the possibly gapped future queue.
type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}







// txList is a "list" of transactions belonging to an account, sorted by account
// nonce. The same type can be used both for storing contiguous transactions for
// the executable/pending queue; and for storing gapped transactions for the non-
// executable/future queue.
type txList struct {
	strict bool         // Whether nonces are strictly continuous or not
	txs    *txSortedMap // Heap indexed sorted hash map of the transactions

	lastStkCheck uint64   // Check all staking transaction validation every 10 blocks
}
// newTxList create a new transaction list for maintaining nonce-indexable fast,
// gapped, sortable transaction lists.
func newTxList(strict bool) *txList {
	return &txList{
		strict:  strict,
		txs:     newTxSortedMap(),
	}
}
// Add tries to insert a new transaction into the list.
func (l *txList) Add(tx model.PoolTransaction) error {
	// If there's an older transaction, abort
	old := l.txs.Get(tx.Nonce())
	if old != nil {
		logrus.Warn("Older transaction is exist. Check it out. Something wrong", spew.Sdump(old))
		//TODO return error
	}
	l.txs.Put(tx)

	return nil
}
// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (l *txList) Ready(start uint64) model.PoolTransactions {
	return l.txs.Ready(start)
}
// Len returns the length of the transaction list.
func (l *txList) Len() int {
	return l.txs.Len()
}
// Empty returns whether the list of transactions is empty or not.
func (l *txList) Empty() bool {
	return l.Len() == 0
}
// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (l *txList) Flatten() model.PoolTransactions {
	return l.txs.Flatten()
}








// txSortedMap is a nonce->transaction hash map with a heap based index to allow
// iterating over the contents in a nonce-incrementing way.
type txSortedMap struct {
	items map[uint64]model.PoolTransaction // Hash map storing the transaction data
	index *nonceHeap                       // Heap of nonces of all the stored transactions (non-strict mode)
	cache model.PoolTransactions           // Cache of the transactions already sorted
}
// newTxSortedMap creates a new nonce-sorted transaction map.
func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]model.PoolTransaction),
		index: new(nonceHeap),
	}
}
// Get retrieves the current transactions associated with the given nonce.
func (m *txSortedMap) Get(nonce uint64) model.PoolTransaction {
	return m.items[nonce]
}
// Put inserts a new transaction into the map, also updating the map's nonce
// index. If a transaction already exists with the same nonce, it's overwritten.
func (m *txSortedMap) Put(tx model.PoolTransaction) {
	nonce := tx.Nonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (m *txSortedMap) Ready(start uint64) model.PoolTransactions {
	// Short circuit if no transactions are available
	if m.index.Len() == 0 || (*m.index)[0] > start {
		return nil
	}
	// Otherwise, start accumulating incremental transactions
	var ready model.PoolTransactions
	for next := (*m.index)[0]; m.index.Len() > 0 && (*m.index)[0] == next; next++ {
		ready = append(ready, m.items[next])
		delete(m.items, next)
		heap.Pop(m.index)
	}
	m.cache = nil

	return ready
}
// Len returns the length of the transaction map.
func (m *txSortedMap) Len() int {
	return len(m.items)
}
// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (m *txSortedMap) Flatten() model.PoolTransactions {
	// If the sorting was not cached yet, create and cache it
	if m.cache == nil {
		m.cache = make(model.PoolTransactions, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(model.PoolTxByNonce(m.cache))
	}
	// Copy the cache to prevent accidental modifications
	txs := make(model.PoolTransactions, len(m.cache))
	copy(txs, m.cache)
	return txs
}
