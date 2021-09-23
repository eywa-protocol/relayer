package block

import "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/blockchain/core/transaction/api"

type Block struct {}

var  (
	txApi = api.NewTransactionApi()
)

func (receiver *Block)   processBlock()  {
	txApi.SaveTx(nil)
}
