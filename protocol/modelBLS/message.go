package modelBLS

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
	Signature []byte
	Mask      []byte
}

type MessageInterface interface {
	MessageToBytes(sig MessageWithSig) *[]byte
	BytesToModelMessage([]byte) *MessageWithSig
}
