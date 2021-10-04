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
)

type MessageWithSig struct {
	Source    int     // NodeID of message's source
	Step      int     // Time step of message
	MsgType   MsgType // Type of message
	History   []MessageWithSig
	Signature common.BlsSignature
	Mask      big.Int
}

type MessageInterface interface {
	MessageToBytes(sig MessageWithSig) *[]byte
	BytesToModelMessage([]byte) *MessageWithSig
}
