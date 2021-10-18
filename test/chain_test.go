package test

import (
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	_common "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	_config "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/log"
	_genesis "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/genesis"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/ledgerstore"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"os"
	"testing"
)

var testBlockStore *ledgerstore.BlockStore
var testStateStore *ledgerstore.StateStore
var testLedgerStore *ledgerstore.LedgerStoreImp

func TestMain(m *testing.M) {
	log.InitLog(0)

	var err error
	testLedgerStore, err = ledgerstore.NewLedgerStore("test/ledger")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewLedgerStore error %s\n", err)
		return
	}

	testBlockDir := "test/block"
	testBlockStore, err = ledgerstore.NewBlockStore(testBlockDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewBlockStore error %s\n", err)
		return
	}
	testStateDir := "test/state"
	merklePath := "test/" + ledgerstore.MerkleTreeStorePath
	testStateStore, err = ledgerstore.NewStateStore(testStateDir, merklePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewStateStore error %s\n", err)
		return
	}
	m.Run()
	err = testLedgerStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testLedgerStore.Close error %s\n", err)
		return
	}
	err = testBlockStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testBlockStore.Close error %s\n", err)
		return
	}
	err = testStateStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testStateStore.Close error %s", err)
		return
	}
	err = os.RemoveAll("./test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.RemoveAll error %s\n", err)
		return
	}
	os.RemoveAll("ActorLog")
}

func TestGenesisBlockInit(t *testing.T) {
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	conf := &_config.GenesisConfig{}
	block, err := _genesis.BuildGenesisBlock([]keypair.PublicKey{pub}, conf)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.NotEqual(t, block.Header.TransactionsRoot, _common.UINT256_EMPTY)
}

func TestInitLedgerStoreWithGenesisBlock(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	acc3 := account.NewAccount("")
	acc4 := account.NewAccount("")
	acc5 := account.NewAccount("")
	acc6 := account.NewAccount("")
	acc7 := account.NewAccount("")

	bookkeepers := []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey, acc4.PublicKey, acc5.PublicKey, acc6.PublicKey, acc7.PublicKey}
	bookkeeper, err := types.AddressFromBookkeepers(bookkeepers)
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	if bookkeeper == _common.ADDRESS_EMPTY {
		t.Errorf("AddressFromBookkeepers error %s", fmt.Errorf("empty address %v", bookkeeper.ToHexString()))
		return
	}
	genesisConfig := _config.DefConfig.Genesis
	block, err := _genesis.BuildGenesisBlock(bookkeepers, genesisConfig)
	//header := &types.Header{
	//	Version:          0,
	//	PrevBlockHash:    common.Uint256{},
	//	TransactionsRoot: common.Uint256{},
	//	Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
	//	Height:           uint32(0),
	//	ConsensusData:    1234567890,
	//	NextBookkeeper:   bookkeeper,
	//}
	//block.Header = header
	//block := &types.Block{
	//	Header:       header,
	//	Transactions: []*types.Transaction{},
	//}

	err = testLedgerStore.InitLedgerStoreWithGenesisBlock(block, bookkeepers)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock error %s", err)
		return
	}

	curBlockHeight := testLedgerStore.GetCurrentBlockHeight()
	curBlockHash := testLedgerStore.GetCurrentBlockHash()
	if curBlockHeight != block.Header.Height {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHeight %d != %d", curBlockHeight, block.Header.Height)
		return
	}
	if curBlockHash != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHash %x != %x", curBlockHash, block.Hash())
		return
	}
	block1, err := testLedgerStore.GetBlockByHeight(curBlockHeight)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed GetBlockByHeight error %s", err)
		return
	}

	if block1.Hash() != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed blockhash %x != %x", block1.Hash(), block.Hash())
		return
	}

	blockByHash, err := testLedgerStore.GetBlockByHash(curBlockHash)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed GetBlockByHash error %s", err)
		return
	}

	if blockByHash.Hash() != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed blockhash %x != %x", blockByHash.Hash(), block.Hash())
		return
	}
}
