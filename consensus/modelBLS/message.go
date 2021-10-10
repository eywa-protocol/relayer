package modelBLS

import (
	"math/big"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

type MsgType int

const (
	Raw = iota
	Ack
	Wit
	Catchup
	BlsSetupPhase
	BlsMembershipKeysParts
)

type Body struct {
	Step       int     // Time step of message
	ActionRoot big.Int // Merkle root of actions in this block
}

type MessageWithSig struct {
	Body
	Source             int     // NodeID of message's source
	MsgType            MsgType // Type of message
	History            []MessageWithSig
	Signature          common.BlsSignature   // Aggregated signature // TODO
	Mask               big.Int               // Bitmask of those who signed
	PublicKey          common.BlsPublicKey   // Aggregated public key of those who signed
	MembershipKeyParts []common.BlsSignature // Used only in BlsMembershipKeysParts msg
}

type MessageInterface interface {
	MessageToBytes(sig MessageWithSig) *[]byte
	BytesToModelMessage([]byte) *MessageWithSig
}
