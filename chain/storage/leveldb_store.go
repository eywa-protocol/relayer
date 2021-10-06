package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBOptions configuration for LevelDB.
type LevelDBOptions struct {
	DataDirectoryPath string `yaml:"DataDirectoryPath"`
}

// LevelDBStore is the official storage implementation for storing and retrieving
// blockchain data.
type LevelDBStore struct {
	db *leveldb.DB
}

//// NewLevelDBStore returns a Store implementation using the boltdb storage engine.
//func NewLevelDBStore(folder string, opts *bolt.Options) (chain.Store, error) {
//	dbPath := path.Join(folder, BoltFileName)
//	db, err := bolt.Open(dbPath, 0660, opts)
//	if err != nil {
//		return nil, err
//	}
//	// create the bucket already
//	err = db.Update(func(tx *bolt.Tx) error {
//		_, err := tx.CreateBucketIfNotExists(blockBucket)
//		if err != nil {
//			return err
//		}
//		return nil
//	})
//
//	return &boltStore{
//		db: db,
//	}, err
//}

// NewLevelDBStore returns a new LevelDBStore object that will
// initialize the database found at the given path.
func NewLevelDBStore(cfg LevelDBOptions) (Store, error) {
	var opts = new(opt.Options) // should be exposed via LevelDBOptions if anything needed

	opts.Filter = filter.NewBloomFilter(10)
	db, err := leveldb.OpenFile(cfg.DataDirectoryPath, opts)
	if err != nil {
		return nil, err
	}

	return &LevelDBStore{db: db}, nil
}

// Put implements the Store interface.
func (s *LevelDBStore) Put(key, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *LevelDBStore) Has(value []byte) (bool, error) {
	return true, nil
}

// Get implements the Store interface.
func (s *LevelDBStore) Get(key []byte) ([]byte, error) {
	value, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		err = ErrKeyNotFound
	}
	return value, err
}

// Delete implements the Store interface.
func (s *LevelDBStore) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

// PutBatch implements the Store interface.
func (s *LevelDBStore) PutBatch(batch Batch) error {
	lvldbBatch := batch.(*leveldb.Batch)
	return s.db.Write(lvldbBatch, nil)
}

// PutChangeSet implements the Store interface.
func (s *LevelDBStore) PutChangeSet(puts map[string][]byte, dels map[string]bool) error {
	tx, err := s.db.OpenTransaction()
	if err != nil {
		return err
	}
	for k := range puts {
		err = tx.Put([]byte(k), puts[k], nil)
		if err != nil {
			tx.Discard()
			return err
		}
	}
	for k := range dels {
		err = tx.Delete([]byte(k), nil)
		if err != nil {
			tx.Discard()
			return err
		}
	}
	return tx.Commit()
}

// Seek implements the Store interface.
func (s *LevelDBStore) Seek(key []byte, f func(k, v []byte)) {
	iter := s.db.NewIterator(util.BytesPrefix(key), nil)
	for iter.Next() {
		f(iter.Key(), iter.Value())
	}
	iter.Release()
}

// Batch implements the Batch interface and returns a leveldb
// compatible Batch.
func (s *LevelDBStore) Batch() Batch {
	return new(leveldb.Batch)
}

// Close implements the Store interface.
func (s *LevelDBStore) Close() error {
	return s.db.Close()
}
