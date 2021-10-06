package chain

import (
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
	coinbasetypes.Transaction := NewCoinbaseTX(genesisEpoch.Serialize())
	genesis := NewGenesisBlock(*header, []*types.Transaction{coinbasetypes.Transaction})
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
			b := tx.Bucket([]byte(types.BlocksBucket))
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
		b := tx.Bucket([]byte(types.BlocksBucket))
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
func (bc *Chain) AddBlock(types.Block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(types.BlocksBucket))
		blockInDb := b.Get(types.Block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(types.Block.Hash, blockData)
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

// Findtypes.Transaction finds a types.Transaction by its ID
/*func (bc *Chain) Findtypes.Transaction(ID []byte) (types.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.types.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(types.Block.PrevBlockHash) == 0 {
			break
		}
	}

	return types.Transaction{}, errors.New("types.Transaction is not found")
}*/
/*
// Iterator returns a BlockchainIterat
func (bc *Chain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.currrentBlockHash, bc.db}

	return bci
}*/



/*// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Chain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(types.Blocks, block.Hash)

		if len(types.Block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}
*/
/*// MineBlock mines a new block with the provided types.Transactions
func (bc *Chain) DelegateBlock(types.Transactions []*types.Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range types.Transactions {
		// TODO: ignore types.Transaction if it's not valid
		if bc.Verifytypes.Transaction(tx) != true {
			log.Panic("ERROR: Invalid types.Transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(types.BlocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(types.BlockData)

		lastHeight = block.Number

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock( header , lastHeight+1, types.Transactions, lastHash, []byte(""), []byte(""))

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(types.BlocksBucket))
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
//// Signtypes.Transaction signs inputs of a types.Transaction
//func (bc *Chain) SignTransaction(tx *types.Transaction, privKey ecdsa.PrivateKey) {
//	prevTXs := make(map[string]types.Transaction)
//	tx.Sign(privKey, prevTXs)
//}
//
//// Verifytypes.Transaction verifies types.Transaction input signatures
//func (bc *Chain) VerifyTransaction(tx *types.Transaction) bool {
//	if tx.IsCoinbase() {
//		return true
//	}
//	//TODO verify txs
//	return true
//}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
