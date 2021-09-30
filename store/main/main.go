package main

import (
	"context"
	"math/big"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
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
	logrus.SetLevel(logrus.Level(4))
	// create mock blockchain
	bc := &testBlockChain{}
	// create tx pool (event pool). Subscribe on event ChainHeadEvent. Started eventLoop.
	txPool := store.NewTxPool(bc)
	defer txPool.Stop()


	// emulate event OracleRequest during 50 sec
	checkClientTimer := time.NewTicker(10 * time.Second)
	go func(ch chan <-store.NewEventSourceChain)  {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer func() {
			checkClientTimer.Stop()
			cancel()
			time.Sleep(10 * time.Second)
			pending, _ := txPool.Pending()
			logrus.Debugln("Ended...\n", spew.Sdump(pending))
			logrus.Debugln("================================================================")
			pendingByChainId, _ := txPool.PendingByChainId()
			for k, txs := range pendingByChainId {
				logrus.Info("Key: ", k)
				for _, tx := range txs {
					adr, _ := tx.SenderAddress()
					logrus.Infof("ChainId: %d Nonce: %d Sender: %s",tx.ChainId(), tx.Nonce(), adr.String())
				}
			}
			logrus.Debugln("Ended 2...\n", spew.Sdump(pendingByChainId))
			os.Exit(0)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-checkClientTimer.C:
				logrus.Info("Sended mock OracleRequest...")
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge: common.HexToAddress("0x376c47978271565f56DEB45495afa69E59c16Ab2"),
						Chainid: big.NewInt(4),
					},
				}
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge: common.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
						Chainid: big.NewInt(4),
					},
				}
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge: common.HexToAddress("0x376c47978271565f56DEB45495afa69E59c16Ab2"),
						Chainid: big.NewInt(4),
					},
				}
				ch <- store.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge: common.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
						Chainid: big.NewInt(94),
					},
				}

			}
		}
	}(txPool.EventSourceReq)
}
