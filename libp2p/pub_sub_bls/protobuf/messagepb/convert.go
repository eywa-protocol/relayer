package messagepb

import (
	"fmt"
	"github.com/DigiU-Lab/p2p-bridge/libp2p/pub_sub_bls/model"
	"github.com/golang/protobuf/proto"
)

type Convert struct{}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg model.Message) (message *PbMessage) {
	source := int64(msg.Source)
	step := int64(msg.Step)

	msgType := MsgType(int(msg.MsgType))

	history := make([]*PbMessage, 0)

	for _, hist := range msg.History {
		history = append(history, convertModelMessage(hist))
	}

	message = &PbMessage{
		Source:  &source,
		Step:    &step,
		MsgType: &msgType,
		History: history,
	}
	return
}

func (c *Convert) MessageToBytes(msg model.Message) *[]byte {
	msgBytes, err := proto.Marshal(convertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	return &msgBytes
}

// ConvertPbMessage is for converting protobuf message to message used in model
func convertPbMessage(msg *PbMessage) (message model.Message) {
	history := make([]model.Message, 0)

	for _, hist := range msg.History {
		history = append(history, convertPbMessage(hist))
	}

	message = model.Message{
		Source:  int(msg.GetSource()),
		Step:    int(msg.GetStep()),
		MsgType: model.MsgType(int(msg.GetMsgType())),
		History: history,
	}
	return
}

func (c *Convert) BytesToModelMessage(msgBytes []byte) *model.Message {
	var pbMessage PbMessage
	err := proto.Unmarshal(msgBytes, &pbMessage)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	modelMsg := convertPbMessage(&pbMessage)
	return &modelMsg
}
