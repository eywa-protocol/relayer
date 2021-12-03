package session

import (
	"math/big"

	"github.com/eywa-protocol/bls-crypto/bls"
)

type MsgType int

const (
	Announce = iota
	Prepare
	Prepared
)

type Body struct {
	BridgeEventHash string
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

type MessageInterface interface {
	MessageToBytes(sig MessageWithSig) *[]byte
	BytesToModelMessage([]byte) *MessageWithSig
}
