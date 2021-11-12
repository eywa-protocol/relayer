package message

import (
	"fmt"
	"google.golang.org/protobuf/runtime/protoimpl"
	"math/big"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/model"
)

type Convert struct{}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg model.MessageWithSig) (message *ConsensusRequest) {
	source := int64(msg.Source)
	step := int64(msg.Step)

	msgType := MsgType(int(msg.MsgType))

	history := make([]*ConsensusRequest, 0)

	for _, hist := range msg.History {
		history = append(history, convertModelMessage(hist))
	}
	message = &ConsensusRequest{
		state:           protoimpl.MessageState{},
		sizeCache:       0,
		unknownFields:   nil,
		MsgType:         &msgType,
		Source:          &source,
		Step:            &step,
		History:         history,
		Signature:       msg.Signature.Marshal(),
		Mask:            msg.Mask.Bytes(),
		PublicKey:       msg.PublicKey.Marshal(),
		BridgeEventHash: &msg.BridgeEventHash,
	}
	return
}

func (c *Convert) MessageToBytes(msg model.MessageWithSig) *[]byte {
	msgBytes, err := proto.Marshal(convertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	return &msgBytes
}

// ConvertConsensusRequestSig is for converting protobuf message to message used in model
func convertConsensusRequestSig(msg *ConsensusRequest) (message model.MessageWithSig) {
	history := make([]model.MessageWithSig, 0)

	for _, hist := range msg.History {
		history = append(history, convertConsensusRequestSig(hist))
	}

	sig, err := bls.UnmarshalSignature(msg.Signature)
	if err != nil {
		logrus.Trace("UnmarshalBlsSignature error: ", err.Error(), msg.Signature)
	}

	pub, err := bls.UnmarshalPublicKey(msg.PublicKey)
	if err != nil {
		logrus.Trace("UnmarshalBlsPublicKey error: ", err.Error(), msg.PublicKey)
	}

	message = model.MessageWithSig{
		Header:    model.Header{int(msg.GetSource()), model.MsgType(msg.GetMsgType())},
		Body:      model.Body{int(msg.GetStep()), msg.GetBridgeEventHash()},
		History:   history,
		Signature: sig,
		Mask:      *new(big.Int).SetBytes(msg.Mask),
		PublicKey: pub,
	}
	return
}

func (c *Convert) BytesToModelMessage(msgBytes []byte) *model.MessageWithSig {
	var ConsensusRequestSig ConsensusRequest
	err := proto.Unmarshal(msgBytes, &ConsensusRequestSig)
	if err != nil {
		logrus.Error("Unmarshal MessageWithSig ", err, string(msgBytes))
		return nil
	}

	modelMsg := convertConsensusRequestSig(&ConsensusRequestSig)
	return &modelMsg
}
