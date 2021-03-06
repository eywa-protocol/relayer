package avoid_importcycle_tmpdir

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain/rawdb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain/storage"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var (
	genesisEpoch        = chain.CreateGenesisEpoch()
	coinbaseTransaction = chain.NewCoinbaseTX(genesisEpoch.Serialize())
	header              = chain.NewHeader(*genesisEpoch)
)

type testBlockChain struct {
	db            storage.Store
	currentBlock  *chain.Block
	chainHeadFeed event.Feed
	scope         event.SubscriptionScope
}

func (bc *testBlockChain) CurrentBlock() *chain.Block {
	return nil
}
func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- chain.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}
func (bc *testBlockChain) StateAt(root common.Hash) error {
	return nil
	//return state.New(root, bc.stateCache)
}

func testTxPool(t *testing.T) {
	blockchain := CreateMockBlockchain(t)
	//RunMockTxPool()
	_ = blockchain
}

func CreateMockBlockchain(t *testing.T) *testBlockChain {
	store, _ := storage.NewLevelDBStore(storage.LevelDBOptions{DataDirectoryPath: t.TempDir()})
	genesisBlock := chain.NewGenesisBlock(*header, []chain.Transaction{*coinbaseTransaction})
	bc := &testBlockChain{
		db:           store,
		currentBlock: genesisBlock,

	}
	//1. added genesis
	rawdb.WriteBlock(bc.db, genesisBlock.Hash, genesisBlock.Number, *genesisBlock)
	rawdb.WriteTxLookupEntries(bc.db, genesisBlock)
	readBlock := rawdb.ReadBlock(bc.db, genesisBlock.Hash, genesisBlock.Number)
	_, blockHash, _, _ := rawdb.ReadTransaction(bc.db, coinbaseTransaction.Hash())

	require.Equal(t, readBlock.Hash, blockHash)

	//2. added simple block with tx
    //todo
	return bc
}

func RunMockTxPool() {
	logrus.SetLevel(logrus.Level(4))
	// create mock blockchain
	bc := &testBlockChain{}
	// create tx pool (event pool). Subscribe on event ChainHeadEvent. Started eventLoop.
	txPool := chain.NewTxPool(bc)
	defer txPool.Stop()

	// emulate event OracleRequest during 50 sec
	checkClientTimer := time.NewTicker(10 * time.Second)
	go func(ch chan<- chain.NewEventSourceChain) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer func() {
			checkClientTimer.Stop()
			cancel()
			time.Sleep(3 * time.Second)
			pending, _ := txPool.Pending()
			logrus.Debugln("Ended...\n", spew.Sdump(pending))
			logrus.Debugln("================================================================")
			pendingByChainId, _ := txPool.PendingByChainId()
			for k, txs := range pendingByChainId {
				logrus.Info("Key: ", k)
				for _, tx := range txs {
					adr := tx.SenderAddress()
					logrus.Infof("ChainId: %d Nonce: %d Sender: %s", tx.ChainId(), tx.Nonce(), adr.String())
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
				ch <- chain.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge:      common.HexToAddress("0x376c47978271565f56DEB45495afa69E59c16Ab2"),
						Chainid:     big.NewInt(4),
					},
				}
				ch <- chain.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge:      common.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
						Chainid:     big.NewInt(4),
					},
				}
				ch <- chain.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge:      common.HexToAddress("0x376c47978271565f56DEB45495afa69E59c16Ab2"),
						Chainid:     big.NewInt(4),
					},
				}
				ch <- chain.NewEventSourceChain{
					EventSourceChain: &wrappers.BridgeOracleRequest{
						RequestType: "setRequest",
						Bridge:      common.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
						Chainid:     big.NewInt(94),
					},
				}

			}
		}
	}(txPool.EventSourceReq)
}
