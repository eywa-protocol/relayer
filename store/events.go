package store

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/store/model"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

// ChainHeadEvent is the struct of chain head event.
type ChainHeadEvent struct{ Block *model.Block }

// NewEventSourceChain the struct of event from source chain
type NewEventSourceChain struct{ EventSourceChain *wrappers.BridgeOracleRequest }
