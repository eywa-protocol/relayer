package test

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	_common "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	_config "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	_genesis "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/genesis"
	"testing"
)

func TestGenesisBlockInit(t *testing.T) {
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	conf := &_config.GenesisConfig{}
	block, err := _genesis.BuildGenesisBlock([]keypair.PublicKey{pub}, conf)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.NotEqual(t, block.Header.TransactionsRoot, _common.UINT256_EMPTY)
}
