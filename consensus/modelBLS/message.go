package modelBLS

import (
	"math/big"

	"gitlab.digiu.ai/blockchainlaboratory/bls-crypto/bls"
)

type MsgType int

const (
	Raw = iota
	Ack
	Wit
	Catchup
	BlsSetupPhase
	BlsSetupParts
)

type Body struct {
	Step       int     // Time step of message
	ActionRoot big.Int // Merkle root of actions in this block
}

type Header struct {
	Source  int     // NodeID of message's source
	MsgType MsgType // Type of message
}

type MessageWithSig struct {
	Header
	Body
	History   []MessageWithSig
	Signature bls.Signature // Aggregated signature
	Mask      big.Int       // Bitmask of those who signed
	PublicKey bls.PublicKey // Aggregated public key of those who signed
}

type MessageBlsSetup struct {
	Header
	MembershipKeyParts []bls.Signature
}

type MessageInterface interface {
	MessageToBytes(sig MessageWithSig) *[]byte
	BytesToModelMessage([]byte) *MessageWithSig
}
