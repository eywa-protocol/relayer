package services

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/entity"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/repository"
)

/**
    get some logic from interface (controller). doing this and  forward other logic on layer domain
*/

type ITransactionService interface {
	Save(tx *entity.Transaction) bool
}
type TransactionService struct {
	TxRepository repository.ITransactionRepository
}




//Save adding row into db (simple example)
func (s *TransactionService) Save(tx *entity.Transaction) bool {
	return s.TxRepository.SaveTx(tx)
}
