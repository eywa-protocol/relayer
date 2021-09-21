package blockchain

import "testing"

func TestNewBlockchain(t *testing.T) {
	bc := CreateBlockchain()
	defer bc.db.Close()
	bc.FindTransaction([]byte("0"))
}
