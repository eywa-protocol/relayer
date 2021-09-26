package main

import (
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/model"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type testBlockChain struct {
	currentBlock     *model.Block
	chainHeadFeed   event.Feed
	scope                  event.SubscriptionScope
}
func (bc *testBlockChain) CurrentBlock() *model.Block {
	return nil
}
func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- store.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}


func main()  {
	logrus.SetLevel(logrus.Level(5))
	// create mock blockchain
	bc := &testBlockChain{}
	// create tx pool (event pool). Subscribe on event ChainHeadEvent. Started eventLoop.
	txPool := store.NewTxPool(bc)
	defer txPool.Stop()
	// emulate event OracleRequest
	checkClientTimer := time.NewTicker(10 * time.Second)
	go func(ch chan <-store.NewEventSourceChain)  {
		defer checkClientTimer.Stop()
		for {
			select {
			case <-checkClientTimer.C:
				logrus.Info("Sended mock OracleRequest...")
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{},
				}
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{},
				}

			}
		}
	}(txPool.EventSourceReq)

}
