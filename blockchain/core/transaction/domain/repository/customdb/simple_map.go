package customdb

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/domain/entity"
)

type Repo struct {

}

func (r Repo) SaveTx(transaction *entity.Transaction) bool {
	panic("implement me")
}

func (r Repo) GetTx(u uint64) (*entity.Transaction, error) {
	panic("implement me")
}

