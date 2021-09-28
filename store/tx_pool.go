package store

import (
   _ "math/big"
   "sync"
   _ "time"

   "github.com/davecgh/go-spew/spew"
   "github.com/ethereum/go-ethereum/common"
   "github.com/sirupsen/logrus"
   "gitlab.digiu.ai/blockchainlaboratory/wrappers"

   "github.com/ethereum/go-ethereum/event"
   "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/model"
)

const (
   chainHeadChanSize = 10
)

// blockChain provides the state of blockchain for
// some pre checks in tx pool and event subscribers.
type blockChain interface {
   CurrentBlock() *model.Block
   SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}

type TxPool struct {
   nonce               uint64
   chain               blockChain
   chainHeadCh         chan ChainHeadEvent
   EventSourceReq      chan NewEventSourceChain
   chainHeadSub        event.Subscription
   scope               event.SubscriptionScope
   pending             map[common.Address]*txList   // All currently processable transactions
   queue               map[common.Address]*txList   // Queued but non-processable transactions
   mu    sync.RWMutex
   wg    sync.WaitGroup // for shutdown sync

   //txErrorSink *types.TransactionErrorSink // All failed txs gets reported here
}

func NewTxPool(chain blockChain) *TxPool {
   pool := &TxPool{
      EventSourceReq:  make(chan NewEventSourceChain, 10),
      chainHeadCh:     make(chan ChainHeadEvent, chainHeadChanSize),
      chain:           chain,
      pending:         make(map[common.Address]*txList),
      queue:           make(map[common.Address]*txList),
      nonce:           1,
   }

   pool.wg.Add(1)
   go pool.loop()
   // Subscribe events from blockchain
   pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)

   return  pool
}

// loop is the event pool's main event loop, waiting for and reacting to
//  events from blockchain as well as for various reporting and  eviction events.
func (pool *TxPool) loop() {
   defer pool.wg.Done()

   // Track the previous head headers for transaction reorgs
   head := pool.chain.CurrentBlock()

   // Keep waiting for and reacting to the various events
   for {
      select {
         // Handle ChainHeadEvent
         case ev := <-pool.chainHeadCh:
            if ev.Block != nil {
               pool.mu.Lock()
                  pool.reset(head.Header(), ev.Block.Header())
                  head = ev.Block
               pool.mu.Unlock()
            }
         // Handle NewEventSourceChain
         case ev := <-pool.EventSourceReq:
            logrus.Info("Got OracleRequest event ")
            logrus.Debug("Got OracleRequest event:", spew.Sdump(ev.EventSourceChain))
            if ev.EventSourceChain != nil {
               pool.mu.Lock()
                  pool.AddEvent(ev.EventSourceChain)
               pool.mu.Unlock()
            }
      }
   }
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (pool *TxPool) reset(oldHead, newHead *model.Header) {

}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
   // Unsubscribe all subscriptions registered from txpool
   pool.scope.Close()

   // Unsubscribe subscriptions registered from blockchain
   pool.chainHeadSub.Unsubscribe()
   pool.wg.Wait()

}
// AddEvent enqueues a single event into the pool if it is valid.
func (pool *TxPool) AddEvent(ev *wrappers.BridgeOracleRequest) {

   // check isExist in inbound events (check double)
   // check. event source address isTrue
   // check chainId 'from to'
   // process nonce by event source address (for double spending)
   // send event to another for sync pool
   // check current block for db == currentBlock subscribe
   // put to store
   tx := model.NewTransaction(ev, pool.nonce)
   poolTxs := model.PoolTransactions{tx}
   pool.addEvLocked(poolTxs)
   //NOTE: At this moment, nonce shared between all addresses
   pool.nonce++
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
/*func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
   return pool.scope.Track(pool.txFeed.Subscribe(ch))
}*/

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) stats() (int, int) {
   return 0,0
}

// addEv attempts to queue a batch of transactions if they are valid.
/*func (pool *TxPool) addEv(txs model.PoolTransactions) []error {
   pool.mu.Lock()
   defer pool.mu.Unlock()

   return pool.addEvLocked(txs)
}*/

// addEvLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *TxPool) addEvLocked(txs model.PoolTransactions) []error {
   adrSenders:= map[common.Address]struct{}{}
   errs      := make([]error, txs.Len())

   for _, tx := range txs {
      err := pool.add(tx)
      if err == nil {
         from, _ := tx.SenderAddress()
         adrSenders[from] = struct{}{}
      }
      // Ignore known transaction for tx rebroadcast case.
      //if err != nil && errCause != ErrKnownTransaction
   }
   if len(adrSenders) > 0 {
      adds := make([]common.Address, len(adrSenders))
      i := 0
      for addr := range adrSenders {
         adds[i] = addr
         i++
      }
      pool.promoteExecutables(adds)
   }

   return errs
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution.
func (pool *TxPool) add(tx model.PoolTransaction) error {
   // If the transaction is already known, discard it
   // If the transaction fails basic validation, discard it
   // If the transaction pool is full, discard underpriced transactions
   // New transaction isn't replacing a pending one, push into queue
   from, _ := tx.SenderAddress()
   if pool.queue[from] == nil {
      pool.queue[from] = newTxList(false)
   }
   err := pool.queue[from].Add(tx)
   if err != nil {
      return err
   }
   logrus.Debugln("Transaction added into non-executable queue:", spew.Sdump(tx))

   return nil
}


func (pool *TxPool) GetTxPoolSize() uint64 {
   return uint64(len(pool.pending)) + uint64(len(pool.queue))
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions are deleted.
func (pool *TxPool) promoteExecutables(accounts []common.Address) {

   // Gather all the accounts potentially needing updates
   if accounts == nil {
      accounts = make([]common.Address, len(pool.queue))
      i := 0
      for addr := range pool.queue {
         accounts[i] = addr
         i++
      }
   }

   // Iterate over all accounts and promote any executable transactions
   for _, addr := range accounts {
      list := pool.queue[addr]
      if list == nil {
         //TODO: LOG like event of security
         panic("Error. Address doesn't exist")
      }
      // Drop all transactions that are deemed too old (low nonce)


      // Gather all executable transactions and promote them
      for _, tx := range list.Ready(pool.nonce) {
         if pool.promoteTx(addr, tx) {
            address, _ := tx.SenderAddress()
            logrus.Infof("Promoting queued transaction. Sender: %s, nonce: %d", address.String(), tx.Nonce())
         }
      }
      // Drop all transactions over the allowed limit


      // Delete the entire queue entry if it became empty.
      if list.Empty() {
         delete(pool.queue, addr)
      }

      // If the pending limit is overflown, start equalizing allowances
      // If we've queued more transactions than the hard limit, drop oldest ones

   }
}

// promoteTx adds a transaction to the pending (processable) list of transactions
// and returns whether it was inserted or an older was better.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) promoteTx(addr common.Address, tx model.PoolTransaction) bool {
   // Try to insert the transaction into the pending queue
   if pool.pending[addr] == nil {
      pool.pending[addr] = newTxList(true)
   }
   list := pool.pending[addr]
   err := list.Add(tx)
   if err != nil {
      return false
   }
   logrus.Debugln("Transaction added into pending (processable) list:", spew.Sdump(tx))

   // Set the potentially new pending nonce and notify any subsystems of the new tx
   //pool.beats[addr] = time.Now()
   //pool.pendingState.SetNonce(addr, tx.Nonce()+1)

   return true
}
