package store

import (
   "sync"

   "github.com/davecgh/go-spew/spew"
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
   chain               blockChain
   chainHeadCh         chan ChainHeadEvent
   EventSourceReq      chan NewEventSourceChain
   chainHeadSub        event.Subscription
   scope               event.SubscriptionScope
   mu    sync.RWMutex
   wg    sync.WaitGroup // for shutdown sync
}

func NewTxPool(chain blockChain) *TxPool {
   pool := &TxPool{
      EventSourceReq:  make(chan NewEventSourceChain, 10),
      chainHeadCh:     make(chan ChainHeadEvent, chainHeadChanSize),
      chain:           chain,
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

func (pool *TxPool) AddEvent(ev *wrappers.BridgeOracleRequest) {

   // check isExist in inbound events (check double)
   // check. event source address isTrue
   // check chainId 'from to'
   // process nonce by event source address (for double spending)
   // send event to another for sync pool
   // put to store



}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
/*func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
   return pool.scope.Track(pool.txFeed.Subscribe(ch))
}*/
