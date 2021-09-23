package db

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLevelDBConnection(t *testing.T) {
	db, err := NewLevelDBDatabase(".test", 1024, 128, false)
	if err != nil {
		t.Errorf("DB is not created, received %v", err)
	} else {
		t.Logf("DB is created, %v", db)
	}

	err = db.Put([]byte("hello"), []byte("world"))
	if err != nil {
		t.Errorf("String is not writed, received %v", err)
	}

	res, err := db.Get([]byte("hello"))
	if err != nil {
		t.Errorf("String is not writed, received %v", err)
	}

	if string(res) != "world" {
		t.Errorf("Unexpected string from db, received %v", string(res))
	} else {
		t.Logf("Data is received, %v", res)
	}
}

func TestLevelDBWithNaiveIndexing(t *testing.T) {
	db, _ := NewLevelDBDatabase(".test_naive", 1024, 128, false)

	txPrefix := []byte("TX_")
	blPrefix := []byte("BL_")

	txCount := 10
	txPerBlock := 5

	var txs = make([][]byte, 10)
	for txId := 0; txId < 10; txId++ {
		txs[txId] = []byte(fmt.Sprintf("tx_%d", txId))
		db.Put(append(txPrefix, byte(txId)), txs[txId])
	}

	blockCount := (txCount-1)/txPerBlock + 1
	var blocks = make([][]byte, blockCount)
	for blockId := 0; blockId < blockCount; blockId++ {
		var b bytes.Buffer
		for blockTx := 0; blockTx < txPerBlock; blockTx++ {
			txId := blockTx + blockId*txPerBlock
			if txId >= txCount {
				break
			}

			b.WriteByte(byte(txId))
		}

		blocks[blockId] = b.Bytes()
		db.Put(append(blPrefix, byte(blockId)), blocks[blockId])
	}

	res, _ := db.Get(append(blPrefix, byte(1)))
	for _, txId := range res {
		tx, _ := db.Get(append(txPrefix, txId))
		t.Logf("[BLOCK %d] Tx: %s", 1, string(tx))
	}
}
