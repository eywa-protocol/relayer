package blockchain

import (
	"testing"
	"github.com/stretchr/testify/require"
)

var(

)



func TestCreateBlockchain(t *testing.T) {
	bc := CreateBlockchain()
	defer bc.db.Close()

}


func TestNOpenBlockchain(t *testing.T) {
	bc := OpenBlockchain()
	defer bc.db.Close()
	b, err := bc.GetBlock([]byte("wqeqwqw"))
	require.NoError(t, err)
	t.Log(b)
}



func TestFindTransaction(t *testing.T) {
	bc := OpenBlockchain()
	defer bc.db.Close()
	tx, err := bc.FindTransaction([]byte("0"))
	require.NoError(t, err)
	t.Log(tx)
}
