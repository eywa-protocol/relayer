package repository

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/entity"
)

type ITransactionRepository interface {
	SaveTx(*entity.Transaction) bool
	GetTx(uint64) (*entity.Transaction, error)
}




