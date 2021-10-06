package chain

import (
	"bytes"
	"encoding/gob"
	"math/big"
)

type Epoch struct {
	Number big.Int
	Sign   []byte
	Leader []byte
}

func CreateGenesisEpoch() *Epoch {
	return &Epoch{
		Number: *big.NewInt(int64(0)),
		Sign:   []byte("genesis signature"),
		Leader: []byte("leader bytes"),
	}
}

func CreateNewEpoch() *Epoch {
	return &Epoch{
		Number: *big.NewInt(int64(0)),
		Sign:   nil,
		Leader: nil,
	}
}

// Serialize serializes the block
func (e *Epoch) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(e)
	if err != nil {
		panic(err)
	}

	return result.Bytes()
}
