package messageSigpb

import (
	"fmt"
	"math/big"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/modelBLS"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/crypto/bls"
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
		PublicKey: msg.PublicKey.Marshal(),
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

	sig, err := bls.UnmarshalSignature(msg.Signature)
	if err != nil {
		logrus.Trace("UnmarshalBlsSignature error: ", err.Error(), msg.Signature)
	}

	pub, err := bls.UnmarshalPublicKey(msg.PublicKey)
	if err != nil {
		logrus.Trace("UnmarshalBlsPublicKey error: ", err.Error(), msg.PublicKey)
	}

	message = modelBLS.MessageWithSig{
		Header:    modelBLS.Header{int(msg.GetSource()), modelBLS.MsgType(int(msg.GetMsgType()))},
		Body:      modelBLS.Body{int(msg.GetStep()), *big.NewInt(0xCAFEBABE)}, // TODO: add ActionRoot to pb
		History:   history,
		Signature: sig,
		Mask:      *new(big.Int).SetBytes(msg.Mask),
		PublicKey: pub,
	}
	return
}

func (c *Convert) BytesToModelMessage(msgBytes []byte) *modelBLS.MessageWithSig {
	var PbMessageSig PbMessageSig
	err := proto.Unmarshal(msgBytes, &PbMessageSig)
	if err != nil {
		logrus.Error("Unmarshal MessageWithSig ", err, string(msgBytes))
		return nil
	}

	modelMsg := convertPbMessageSig(&PbMessageSig)
	return &modelMsg
}
