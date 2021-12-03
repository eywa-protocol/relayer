package message

import (
	"fmt"
	"math/big"

	"google.golang.org/protobuf/runtime/protoimpl"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/session"
)

type Convert struct{}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg session.MessageWithSig) (message *ConsensusRequest) {
	source := int64(msg.Source)
	msgType := MsgType(msg.MsgType)
	message = &ConsensusRequest{
		state:           protoimpl.MessageState{},
		sizeCache:       0,
		unknownFields:   nil,
		MsgType:         &msgType,
		Source:          &source,
		Signature:       msg.Signature.Marshal(),
		Mask:            msg.Mask.Bytes(),
		PublicKey:       msg.PublicKey.Marshal(),
		BridgeEventHash: &msg.BridgeEventHash,
	}
	return
}

func (c *Convert) MessageToBytes(msg session.MessageWithSig) *[]byte {
	msgBytes, err := proto.Marshal(convertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	return &msgBytes
}

// ConvertConsensusRequestSig is for converting protobuf message to message used in model
func convertConsensusRequestSig(msg *ConsensusRequest) (message session.MessageWithSig) {
	history := make([]session.MessageWithSig, 0)

	sig, err := bls.UnmarshalSignature(msg.Signature)
	if err != nil {
		logrus.Trace("UnmarshalBlsSignature error: ", err.Error(), msg.Signature)
	}

	pub, err := bls.UnmarshalPublicKey(msg.PublicKey)
	if err != nil {
		logrus.Trace("UnmarshalBlsPublicKey error: ", err.Error(), msg.PublicKey)
	}

	message = session.MessageWithSig{
		Header:    session.Header{Source: int(msg.GetSource()), MsgType: session.MsgType(msg.GetMsgType())},
		Body:      session.Body{msg.GetBridgeEventHash()},
		History:   history,
		Signature: sig,
		Mask:      *new(big.Int).SetBytes(msg.Mask),
		PublicKey: pub,
	}
	return
}

func (c *Convert) BytesToModelMessage(msgBytes []byte) *session.MessageWithSig {
	var ConsensusRequestSig ConsensusRequest
	err := proto.Unmarshal(msgBytes, &ConsensusRequestSig)
	if err != nil {
		logrus.Error("Unmarshal MessageWithSig ", err, string(msgBytes))
		return nil
	}

	modelMsg := convertConsensusRequestSig(&ConsensusRequestSig)
	return &modelMsg
}
