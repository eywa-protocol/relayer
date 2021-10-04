package messageSigpb

import (
	"fmt"
	"math/big"

	"github.com/golang/protobuf/proto"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/modelBLS"
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
		Signature: msg.Signature.Marshal(),
		Mask:      msg.Mask.Bytes(),
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

	sig, _ := common.UnmarshalBlsSignature(msg.Signature)
	message = modelBLS.MessageWithSig{
		Source:    int(msg.GetSource()),
		Step:      int(msg.GetStep()),
		MsgType:   modelBLS.MsgType(int(msg.GetMsgType())),
		History:   history,
		Signature: sig,
		Mask:      *new(big.Int).SetBytes(msg.Mask),
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
