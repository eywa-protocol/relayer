package field

// Field names for logrus fields
// if you log error with fields you must get names from here
// if needed name not exists add it here as const and use

const (
	PeerId         = "peerId"
	NodeAddress    = "nodeAddress"
	NodeRendezvous = "nodeRendezvous"

	CainId       = "chainID"
	EthUrl       = "ethUrl"
	EcdsaAddress = "ecdsaAddress"
	Balance      = "balance"

	RequestType    = "requestType"
	BridgeAddress  = "bridgeAddress"
	RequestId      = "requestId"
	ReceiveSide    = "receiveSide"
	OppositeBridge = "oppositeBridge"

	BridgeRequest = "bridgeRequest"
	TxId          = "txID"

	ConsensusRendezvous = "consensusRendezvous"
	ConsensusLeader     = "consensusLeader"
	RequestLeader       = "requestLeader"
)
