package messageSigpb

import (
	"fmt"
	"math/big"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/modelBLS"
)

type Convert struct{}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg modelBLS.MessageWithSig) (message *PbMessageSig) {
	source := int64(msg.Source)
	step := int64(msg.Body.Step)

	msgType := MsgType(int(msg.MsgType))

	history := make([]*PbMessageSig, 0)
	for _, hist := range msg.History {
		history = append(history, convertModelMessage(hist))
	}

	mkp := make([][]byte, 0)
	for _, mki := range msg.MembershipKeyParts {
		mkp = append(mkp, mki.Marshal())
	}

	message = &PbMessageSig{
		Source:             &source,
		Step:               &step,
		MsgType:            &msgType,
		History:            history,
		Signature:          msg.Signature.Marshal(),
		Mask:               msg.Mask.Bytes(),
		PublicKey:          msg.PublicKey.Marshal(),
		PartMembershipKeys: mkp,
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

	sig, err := common.UnmarshalBlsSignature(msg.Signature)
	if err != nil {
		//logrus.Error("UnmarshalBlsSignature error: ", err.Error(), msg.Signature)
	}

	pub, err := common.UnmarshalBlsPublicKey(msg.PublicKey)
	if err != nil {
		//logrus.Error("UnmarshalBlsPublicKey error: ", err.Error(), msg.PublicKey)
	}

	mks := make([]common.BlsSignature, 0)
	for _, raw := range msg.PartMembershipKeys {
		mki, err := common.UnmarshalBlsSignature(raw)
		if err != nil {
			logrus.Error("UnmarshalBlsSignature error: ", err.Error(), mki)
		}
		mks = append(mks, mki)
	}

	message = modelBLS.MessageWithSig{
		Body:               modelBLS.Body{int(msg.GetStep()), *big.NewInt(0xCAFEBABE)}, // TODO: add ActionRoot to pb
		Source:             int(msg.GetSource()),
		MsgType:            modelBLS.MsgType(int(msg.GetMsgType())),
		History:            history,
		Signature:          sig,
		Mask:               *new(big.Int).SetBytes(msg.Mask),
		PublicKey:          pub,
		MembershipKeyParts: mks,
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
