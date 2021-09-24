package blockchainA

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/adapters/leveldb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/adapters/sqldb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/model"
)

type ServiceInterface interface {
	GetCurrentBlock() model.Block
}

type blockchainService struct {
	p leveldb.Repository
	n sqldb.Repository
}

func (b *blockchainService) GetCurrentBlock() *model.Block {
	b.getTx()
}
