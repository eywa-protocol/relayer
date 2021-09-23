package api

import (
     "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/entity"
     "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/repository/customdb"
     "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/services"
)


/**
    layer on edge of interaction with other components like block, state and so on
*/

type TransactionApi struct {
       txService services.ITransactionService
}

func NewTransactionApi()  *TransactionApi {
     repo := customdb.Repo{}
     transactionServiceImpl :=services.TransactionService{
          TxRepository: repo,
     }

     return &TransactionApi{&transactionServiceImpl}
}

// SaveTx save simple transaction in db (only example)
func (t *TransactionApi) SaveTx(tx *entity.Transaction) bool {
     //TODO... some logic for another services like validate, check origin and  final save it
     //TODO... some logic for another services like validate, check origin and  final save it
     //TODO... some logic for another services like validate, check origin and  final save it

     return t.txService.Save(tx)
}


