package chain

import (
	"crypto/ecdsa"
	"errors"
	"log"
	"os"
	//TODO change this to LevelDB
	"github.com/boltdb/bolt"
)

const dbFile = "../.data/blockchain.db"
const blocksBucket = "blocks"

// Chain implements interactions with a DB
type Chain struct {
	genesisEpoch *Epoch
	db           *bolt.DB
}

// Genesis retrieves the chain's genesis block.
func (bc *Chain) Genesis() *Epoch {
	return bc.genesisEpoch
}

/*func CreateOrOpenChainFromFile() (chain *Chain) {
	genesisEpoch := CreateGenesisEpoch()
	header := NewHeader(*genesisEpoch)
	coinbaseTransaction := NewCoinbaseTX(genesisEpoch.Serialize())
	genesis := NewGenesisBlock(*header, []*Transaction{coinbaseTransaction})
	dbFile := fmt.Sprintf(dbFile)

	db, err := NewLevelDBDatabase(dbFile, 1024, 128, false)

	err = db.Put(genesis.Hash.Bytes(), genesis.Serialize())
	if err != nil {
		log.Panic(err)
	}
	return nil
	}

		chain = &Chain{
			genesisEpoch: genesisEpoch,
			db:           db,
		}
		log.Println("Chain already exists.")
	} else {
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			qwe = b.Get(genesis.Hash.Bytes(),)

			return nil
		})
		if err != nil {
			log.Panic(err)
		}
	}
}

		return
	}*/

/*
//// OpenBlockchain opens Blockchain
func OpenBlockchain() *Chain {
	dbFile := fmt.Sprintf(dbFile)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}


	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Chain{db}
	log.Print(bc)
	return &bc
}*/

/*// AddBlock saves the block into the blockchain
func (bc *Chain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Number > lastBlock.Number {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.currrentBlockHash = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}*/

// FindTransaction finds a transaction by its ID
/*func (bc *Chain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}*/
/*
// Iterator returns a BlockchainIterat
func (bc *Chain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.currrentBlockHash, bc.db}

	return bci
}*/

// GetBestHeight returns the height of the latest block
func (bc *Chain) GetBestHeight() uint64 {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Number
}

// GetBlock finds a block by its hash and returns it
func (bc *Chain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

/*// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Chain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}
*/
/*// MineBlock mines a new block with the provided transactions
func (bc *Chain) DelegateBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Number

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock( header , lastHeight+1, transactions, lastHash, []byte(""), []byte(""))

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.currrentBlockHash = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}
*/
// SignTransaction signs inputs of a Transaction
func (bc *Chain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Chain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//TODO verify txs
	return true
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
