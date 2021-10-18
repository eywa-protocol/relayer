package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestLogCache_Exists(t *testing.T) {
	cache := NewLogCache()
	for blockNumber := 1; blockNumber < 4; blockNumber++ {
		for i := 1; i < 4; i++ {
			log := types.Log{
				BlockNumber: uint64(blockNumber),
				TxHash:      common.BigToHash(big.NewInt(int64(blockNumber * i))),
			}
			exists := cache.Exists(log)
			assert.False(t, exists)
			exists = cache.Exists(log)
			assert.True(t, exists)
			assert.Equal(t, 1, len(cache.data))
		}
	}
	blockNumber := 1
	for i := 1; i < 4; i++ {
		log := types.Log{
			BlockNumber: uint64(blockNumber),
			TxHash:      common.BigToHash(big.NewInt(int64(blockNumber * i))),
		}
		exists := cache.Exists(log)
		assert.False(t, exists)
		exists = cache.Exists(log)
		assert.True(t, exists)
		assert.Equal(t, 2, len(cache.data))
	}

	blockNumber = 2
	for i := 1; i < 4; i++ {
		log := types.Log{
			BlockNumber: uint64(blockNumber),
			TxHash:      common.BigToHash(big.NewInt(int64(blockNumber * i))),
		}
		exists := cache.Exists(log)
		assert.False(t, exists)
		exists = cache.Exists(log)
		assert.True(t, exists)
		assert.Equal(t, 3, len(cache.data))
	}

	blockNumber = 3
	for i := 1; i < 4; i++ {
		log := types.Log{
			BlockNumber: uint64(blockNumber),
			TxHash:      common.BigToHash(big.NewInt(int64(blockNumber * i))),
		}
		exists := cache.Exists(log)
		assert.True(t, exists)
		assert.Equal(t, 1, len(cache.data))
	}

	fmt.Println(cache)
}
