// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messageWithSignature.proto

package messageSigpb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type MsgType int32

const (
	MsgType_Raw     MsgType = 0
	MsgType_Ack     MsgType = 1
	MsgType_Wit     MsgType = 2
	MsgType_Catchup MsgType = 3
)

var MsgType_name = map[int32]string{
	0: "Raw",
	1: "Ack",
	2: "Wit",
	3: "Catchup",
}

var MsgType_value = map[string]int32{
	"Raw":     0,
	"Ack":     1,
	"Wit":     2,
	"Catchup": 3,
}

func (x MsgType) Enum() *MsgType {
	p := new(MsgType)
	*p = x
	return p
}

func (x MsgType) String() string {
	return proto.EnumName(MsgType_name, int32(x))
}

func (x *MsgType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(MsgType_value, data, "MsgType")
	if err != nil {
		return err
	}
	*x = MsgType(value)
	return nil
}

func (MsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_617ec186a0c1f899, []int{0}
}

type PbMessageSig struct {
	MsgType              *MsgType        `protobuf:"varint,1,req,name=Msg_type,json=MsgType,enum=messageSigpb.MsgType" json:"Msg_type,omitempty"`
	Source               *int64          `protobuf:"varint,2,req,name=source" json:"source,omitempty"`
	Step                 *int64          `protobuf:"varint,3,req,name=step" json:"step,omitempty"`
	History              []*PbMessageSig `protobuf:"bytes,4,rep,name=history" json:"history,omitempty"`
	Signature            []byte          `protobuf:"bytes,5,opt,name=signature" json:"signature,omitempty"`
	Mask                 []byte          `protobuf:"bytes,6,opt,name=mask" json:"mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PbMessageSig) Reset()         { *m = PbMessageSig{} }
func (m *PbMessageSig) String() string { return proto.CompactTextString(m) }
func (*PbMessageSig) ProtoMessage()    {}
func (*PbMessageSig) Descriptor() ([]byte, []int) {
	return fileDescriptor_617ec186a0c1f899, []int{0}
}

func (m *PbMessageSig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PbMessageSig.Unmarshal(m, b)
}
func (m *PbMessageSig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PbMessageSig.Marshal(b, m, deterministic)
}
func (m *PbMessageSig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PbMessageSig.Merge(m, src)
}
func (m *PbMessageSig) XXX_Size() int {
	return xxx_messageInfo_PbMessageSig.Size(m)
}
func (m *PbMessageSig) XXX_DiscardUnknown() {
	xxx_messageInfo_PbMessageSig.DiscardUnknown(m)
}

var xxx_messageInfo_PbMessageSig proto.InternalMessageInfo

func (m *PbMessageSig) GetMsgType() MsgType {
	if m != nil && m.MsgType != nil {
		return *m.MsgType
	}
	return MsgType_Raw
}

func (m *PbMessageSig) GetSource() int64 {
	if m != nil && m.Source != nil {
		return *m.Source
	}
	return 0
}

func (m *PbMessageSig) GetStep() int64 {
	if m != nil && m.Step != nil {
		return *m.Step
	}
	return 0
}

func (m *PbMessageSig) GetHistory() []*PbMessageSig {
	if m != nil {
		return m.History
	}
	return nil
}

func (m *PbMessageSig) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *PbMessageSig) GetMask() []byte {
	if m != nil {
		return m.Mask
	}
	return nil
}

func init() {
	proto.RegisterEnum("messageSigpb.MsgType", MsgType_name, MsgType_value)
	proto.RegisterType((*PbMessageSig)(nil), "messageSigpb.PbMessageSig")
}

func init() { proto.RegisterFile("messageWithSignature.proto", fileDescriptor_617ec186a0c1f899) }

var fileDescriptor_617ec186a0c1f899 = []byte{
	// 230 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x8f, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0x4d, 0xb6, 0x36, 0x3a, 0x0d, 0x12, 0xe6, 0x50, 0x96, 0xe2, 0x21, 0x78, 0x0a, 0x1e,
	0x02, 0x06, 0x5f, 0x40, 0x3c, 0x07, 0x64, 0x2b, 0xf4, 0x28, 0xdb, 0xb0, 0x6c, 0x96, 0x52, 0xb3,
	0x64, 0x26, 0x48, 0x1e, 0xd4, 0xf7, 0x91, 0x2c, 0x8d, 0x6d, 0x6f, 0xbf, 0xf9, 0xbe, 0xf9, 0xf7,
	0xc1, 0xe6, 0x68, 0x88, 0xb4, 0x35, 0x3b, 0xc7, 0xed, 0xd6, 0xd9, 0x6f, 0xcd, 0x43, 0x6f, 0x4a,
	0xdf, 0x77, 0xdc, 0x61, 0x7a, 0xf2, 0xb6, 0xce, 0xfa, 0xfd, 0xd3, 0x6f, 0x04, 0xe9, 0xc7, 0xbe,
	0xfe, 0x97, 0xf0, 0x05, 0xee, 0x6a, 0xb2, 0x5f, 0x3c, 0x7a, 0x23, 0xa3, 0x3c, 0x2e, 0x1e, 0xaa,
	0x75, 0x79, 0x39, 0x51, 0xce, 0xae, 0x4a, 0x6a, 0xb2, 0x9f, 0xa3, 0x37, 0xb8, 0x86, 0x25, 0x75,
	0x43, 0xdf, 0x18, 0x19, 0xe7, 0x71, 0x21, 0xd4, 0xa9, 0x42, 0x84, 0x05, 0xb1, 0xf1, 0x52, 0x04,
	0x35, 0x30, 0xbe, 0x42, 0xd2, 0x3a, 0xe2, 0xae, 0x1f, 0xe5, 0x22, 0x17, 0xc5, 0xaa, 0xda, 0x5c,
	0x6f, 0xbf, 0xfc, 0x45, 0xcd, 0xad, 0xf8, 0x08, 0xf7, 0x34, 0xc7, 0x90, 0xb7, 0x79, 0x54, 0xa4,
	0xea, 0x2c, 0x4c, 0x77, 0x8e, 0x9a, 0x0e, 0x72, 0x19, 0x8c, 0xc0, 0xcf, 0xd5, 0x39, 0x06, 0x26,
	0x20, 0x94, 0xfe, 0xc9, 0x6e, 0x26, 0x78, 0x6b, 0x0e, 0x59, 0x34, 0xc1, 0xce, 0x71, 0x16, 0xe3,
	0x0a, 0x92, 0x77, 0xcd, 0x4d, 0x3b, 0xf8, 0x4c, 0xfc, 0x05, 0x00, 0x00, 0xff, 0xff, 0xd3, 0x28,
	0x6c, 0x12, 0x36, 0x01, 0x00, 0x00,
}