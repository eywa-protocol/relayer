package messageSigpb

import (
	"fmt"
	"github.com/digiu-ai/p2p-bridge/libp2p/pub_sub_bls/modelBLS"
	"github.com/golang/protobuf/proto"
)

type Convert struct{}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg modelBLS.MessageWithSig) (message *PbMessageSig) {
	source := int64(msg.Source)
	step := int64(msg.Step)

	msgType := MsgType(int(msg.MsgType))

	history := make([]*PbMessageSig, 0)

	for _, hist := range msg.History {
		history = append(history, convertModelMessage(hist))
	}

	message = &PbMessageSig{
		Source:    &source,
		Step:      &step,
		MsgType:   &msgType,
		History:   history,
		Signature: msg.Signature,
		Mask:      msg.Mask,
	}
	return
}

func (c *Convert) MessageToBytes(msg modelBLS.MessageWithSig) *[]byte {
	msgBytes, err := proto.Marshal(convertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	return &msgBytes
}

// ConvertPbMessageSig is for converting protobuf message to message used in model
func convertPbMessageSig(msg *PbMessageSig) (message modelBLS.MessageWithSig) {
	history := make([]modelBLS.MessageWithSig, 0)

	for _, hist := range msg.History {
		history = append(history, convertPbMessageSig(hist))
	}

	message = modelBLS.MessageWithSig{
		Source:    int(msg.GetSource()),
		Step:      int(msg.GetStep()),
		MsgType:   modelBLS.MsgType(int(msg.GetMsgType())),
		History:   history,
		Signature: msg.Signature,
		Mask:      msg.Mask,
	}
	return
}

func (c *Convert) BytesToModelMessage(msgBytes []byte) *modelBLS.MessageWithSig {
	var PbMessageSig PbMessageSig
	err := proto.Unmarshal(msgBytes, &PbMessageSig)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	modelMsg := convertPbMessageSig(&PbMessageSig)
	return &modelMsg
}
