package db

import (
	"io"
)

type KeyValueReader interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
}

type KeyValueWriter interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type KeyValueStore interface {
	KeyValueReader
	KeyValueWriter
	io.Closer
}

// Database contains all the methods required by the high level database to not
// only access the key-value data store but also the chain freezer.
type Database interface {
	KeyValueReader
	KeyValueWriter
	io.Closer
}

// nofreezedb is a database wrapper that disables freezer data retrievals.
type nofreezedb struct {
	KeyValueStore
}

// NewDatabase creates a high level database on top of a given key-value data
// store without a freezer moving immutable chain segments into cold storage.
func NewDatabase(db KeyValueStore) Database {
	return &nofreezedb{KeyValueStore: db}
}

// NewLevelDBDatabase creates a persistent key-value database without a freezer
// moving immutable chain segments into cold storage.
func NewLevelDBDatabase(file string, cache int, handles int, readonly bool) (Database, error) {
	db, err := NewLevelDb(file, cache, handles, readonly)
	if err != nil {
		return nil, err
	}
	return NewDatabase(db), nil
}
