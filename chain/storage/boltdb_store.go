package storage

import (
	"errors"
	"io"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/chain"
	bolt "go.etcd.io/bbolt"
)

// boldStore implements the Store interface using the kv storage boltdb (native
// golang implementation). Internally, Beacons are stored as JSON-encoded in the
// db file.
type boltStore struct {
	sync.Mutex
	db *bolt.DB
}

var blockBucket = []byte("blocks")

// BoltFileName is the name of the file boltdb writes to
const BoltFileName = "relayer.db"

// NewBoltStore returns a Store implementation using the boltdb storage engine.
func NewBoltStore(folder string, opts *bolt.Options) (chain.Store, error) {
	dbPath := path.Join(folder, BoltFileName)
	db, err := bolt.Open(dbPath, 0660, opts)
	if err != nil {
		return nil, err
	}
	// create the bucket already
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(blockBucket)
		if err != nil {
			return err
		}
		return nil
	})

	return &boltStore{
		db: db,
	}, err
}

func (b *boltStore) Len() int {
	var length = 0
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		length = bucket.Stats().KeyN
		return nil
	})
	if err != nil {
		logrus.Warn("boltdb", "error ge	tting length", "err", err)
	}
	return length
}

func (b *boltStore) Close() {
	if err := b.db.Close(); err != nil {
		logrus.Debug("boltdb", "close", "err", err)
	}
}

// Put implements the Store interface. WARNING: It does NOT verify that this
// block is not already saved in the database or not.
func (b *boltStore) Put(block *chain.Block) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		buff, err := block.Marshal()
		if err != nil {
			return err
		}
		return bucket.Put((chain.NumberToBytes(block.Number)), buff)
	})
	if err != nil {
		return err
	}
	return nil
}

// ErrNoBLockSaved is the error returned when no block have been saved in the
// database yet.
var ErrNoBlockSaved = errors.New("block not found in database")

// Last returns the last block signature saved into the db
func (b *boltStore) Last() (*chain.Block, error) {
	var block *chain.Block
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		cursor := bucket.Cursor()
		_, v := cursor.Last()
		if v == nil {
			return ErrNoBlockSaved
		}
		b := &chain.Block{}
		if err := b.Unmarshal(v); err != nil {
			return err
		}
		block = b
		return nil
	})
	return block, err
}

// Get returns the block saved at this round
func (b *boltStore) Get(blockNumber uint64) (*chain.Block, error) {
	var block *chain.Block
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		v := bucket.Get(chain.NumberToBytes(blockNumber))
		if v == nil {
			return ErrNoBlockSaved
		}
		b := &chain.Block{}
		if err := b.Unmarshal(v); err != nil {
			return err
		}
		block = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return block, err
}

func (b *boltStore) Del(round uint64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		return bucket.Delete(chain.NumberToBytes(round))
	})
}

func (b *boltStore) Cursor(fn func(chain.Cursor)) {
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blockBucket)
		c := bucket.Cursor()
		fn(&boltCursor{Cursor: c})
		return nil
	})
	if err != nil {
		logrus.Warn("boltdb", "error getting cursor", "err", err)
	}
}

// SaveTo saves the bolt database to an alternate file.
func (b *boltStore) SaveTo(w io.Writer) error {
	return b.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

type boltCursor struct {
	*bolt.Cursor
}

func (c *boltCursor) First() *chain.Block {
	k, v := c.Cursor.First()
	if k == nil {
		return nil
	}
	b := new(chain.Block)
	if err := b.Unmarshal(v); err != nil {
		return nil
	}
	return b
}

func (c *boltCursor) Next() *chain.Block {
	k, v := c.Cursor.Next()
	if k == nil {
		return nil
	}
	b := new(chain.Block)
	if err := b.Unmarshal(v); err != nil {
		return nil
	}
	return b
}

func (c *boltCursor) Seek(round uint64) *chain.Block {
	k, v := c.Cursor.Seek(chain.NumberToBytes(round))
	if k == nil {
		return nil
	}
	b := new(chain.Block)
	if err := b.Unmarshal(v); err != nil {
		return nil
	}
	return b
}

func (c *boltCursor) Last() *chain.Block {
	k, v := c.Cursor.Last()
	if k == nil {
		return nil
	}
	b := new(chain.Block)
	if err := b.Unmarshal(v); err != nil {
		return nil
	}
	return b
}
