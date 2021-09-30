package chain

import (
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

// ChainHeadEvent is the struct of chain head event.
type ChainHeadEvent struct{ Block *Block }

// NewEventSourceChain the struct of event from source chain
type NewEventSourceChain struct{ EventSourceChain *wrappers.BridgeOracleRequest }
