// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.3
// source: message.proto

package message

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MsgType int32

const (
	MsgType_Announce MsgType = 0
	MsgType_Prepare  MsgType = 1
	MsgType_Prepared MsgType = 2
	MsgType_Commit   MsgType = 3
	MsgType_Commited MsgType = 4
)

// Enum value maps for MsgType.
var (
	MsgType_name = map[int32]string{
		0: "Announce",
		1: "Prepare",
		2: "Prepared",
		3: "Commit",
		4: "Commited",
	}
	MsgType_value = map[string]int32{
		"Announce": 0,
		"Prepare":  1,
		"Prepared": 2,
		"Commit":   3,
		"Commited": 4,
	}
)

func (x MsgType) Enum() *MsgType {
	p := new(MsgType)
	*p = x
	return p
}

func (x MsgType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MsgType) Descriptor() protoreflect.EnumDescriptor {
	return file_message_proto_enumTypes[0].Descriptor()
}

func (MsgType) Type() protoreflect.EnumType {
	return &file_message_proto_enumTypes[0]
}

func (x MsgType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *MsgType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = MsgType(num)
	return nil
}

// Deprecated: Use MsgType.Descriptor instead.
func (MsgType) EnumDescriptor() ([]byte, []int) {
	return file_message_proto_rawDescGZIP(), []int{0}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Request:
	//	*Message_Consensus
	//	*Message_Drand
	//	*Message_Uptime
	Request isMessage_Request `protobuf_oneof:"request"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_message_proto_rawDescGZIP(), []int{0}
}

func (m *Message) GetRequest() isMessage_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (x *Message) GetConsensus() *ConsensusRequest {
	if x, ok := x.GetRequest().(*Message_Consensus); ok {
		return x.Consensus
	}
	return nil
}

func (x *Message) GetDrand() *DrandRequest {
	if x, ok := x.GetRequest().(*Message_Drand); ok {
		return x.Drand
	}
	return nil
}

func (x *Message) GetUptime() *UptimeRequest {
	if x, ok := x.GetRequest().(*Message_Uptime); ok {
		return x.Uptime
	}
	return nil
}

type isMessage_Request interface {
	isMessage_Request()
}

type Message_Consensus struct {
	Consensus *ConsensusRequest `protobuf:"bytes,5,opt,name=consensus,oneof"`
}

type Message_Drand struct {
	Drand *DrandRequest `protobuf:"bytes,6,opt,name=drand,oneof"`
}

type Message_Uptime struct {
	Uptime *UptimeRequest `protobuf:"bytes,7,opt,name=uptime,oneof"`
}

func (*Message_Consensus) isMessage_Request() {}

func (*Message_Drand) isMessage_Request() {}

func (*Message_Uptime) isMessage_Request() {}

type ConsensusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MsgType         *MsgType            `protobuf:"varint,1,req,name=Msg_type,json=MsgType,enum=message.MsgType" json:"Msg_type,omitempty"`
	Source          *int64              `protobuf:"varint,2,req,name=source" json:"source,omitempty"`
	Step            *int64              `protobuf:"varint,3,req,name=step" json:"step,omitempty"`
	History         []*ConsensusRequest `protobuf:"bytes,4,rep,name=history" json:"history,omitempty"`
	Signature       []byte              `protobuf:"bytes,5,opt,name=signature" json:"signature,omitempty"`
	Mask            []byte              `protobuf:"bytes,6,opt,name=mask" json:"mask,omitempty"`
	PublicKey       []byte              `protobuf:"bytes,7,opt,name=public_key,json=publicKey" json:"public_key,omitempty"`
	BridgeEventHash *string             `protobuf:"bytes,8,opt,name=bridgeEventHash" json:"bridgeEventHash,omitempty"`
}

func (x *ConsensusRequest) Reset() {
	*x = ConsensusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConsensusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConsensusRequest) ProtoMessage() {}

func (x *ConsensusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConsensusRequest.ProtoReflect.Descriptor instead.
func (*ConsensusRequest) Descriptor() ([]byte, []int) {
	return file_message_proto_rawDescGZIP(), []int{1}
}

func (x *ConsensusRequest) GetMsgType() MsgType {
	if x != nil && x.MsgType != nil {
		return *x.MsgType
	}
	return MsgType_Announce
}

func (x *ConsensusRequest) GetSource() int64 {
	if x != nil && x.Source != nil {
		return *x.Source
	}
	return 0
}

func (x *ConsensusRequest) GetStep() int64 {
	if x != nil && x.Step != nil {
		return *x.Step
	}
	return 0
}

func (x *ConsensusRequest) GetHistory() []*ConsensusRequest {
	if x != nil {
		return x.History
	}
	return nil
}

func (x *ConsensusRequest) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *ConsensusRequest) GetMask() []byte {
	if x != nil {
		return x.Mask
	}
	return nil
}

func (x *ConsensusRequest) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *ConsensusRequest) GetBridgeEventHash() string {
	if x != nil && x.BridgeEventHash != nil {
		return *x.BridgeEventHash
	}
	return ""
}

type DrandRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DrandRequest) Reset() {
	*x = DrandRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DrandRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DrandRequest) ProtoMessage() {}

func (x *DrandRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DrandRequest.ProtoReflect.Descriptor instead.
func (*DrandRequest) Descriptor() ([]byte, []int) {
	return file_message_proto_rawDescGZIP(), []int{2}
}

type UptimeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UptimeRequest) Reset() {
	*x = UptimeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UptimeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UptimeRequest) ProtoMessage() {}

func (x *UptimeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UptimeRequest.ProtoReflect.Descriptor instead.
func (*UptimeRequest) Descriptor() ([]byte, []int) {
	return file_message_proto_rawDescGZIP(), []int{3}
}

var File_message_proto protoreflect.FileDescriptor

var file_message_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xb0, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x39, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75,
	0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x00, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x12,
	0x2d, 0x0a, 0x05, 0x64, 0x72, 0x61, 0x6e, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x44, 0x72, 0x61, 0x6e, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x05, 0x64, 0x72, 0x61, 0x6e, 0x64, 0x12, 0x30,
	0x0a, 0x06, 0x75, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16,
	0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x06, 0x75, 0x70, 0x74, 0x69, 0x6d, 0x65,
	0x42, 0x09, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x9c, 0x02, 0x0a, 0x10,
	0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x2c, 0x0a, 0x08, 0x4d, 0x73, 0x67, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x02,
	0x28, 0x0e, 0x32, 0x11, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x73, 0x67,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x52, 0x07, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x03, 0x52, 0x06,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x74, 0x65, 0x70, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x03, 0x52, 0x04, 0x73, 0x74, 0x65, 0x70, 0x12, 0x33, 0x0a, 0x07, 0x68, 0x69,
	0x73, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x07, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x12,
	0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x6d, 0x61, 0x73, 0x6b, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x6d, 0x61, 0x73,
	0x6b, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79,
	0x12, 0x28, 0x0a, 0x0f, 0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48,
	0x61, 0x73, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x62, 0x72, 0x69, 0x64, 0x67,
	0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x61, 0x73, 0x68, 0x22, 0x0e, 0x0a, 0x0c, 0x44, 0x72,
	0x61, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0f, 0x0a, 0x0d, 0x55, 0x70,
	0x74, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2a, 0x4d, 0x0a, 0x08, 0x4d,
	0x73, 0x67, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x41, 0x6e, 0x6e, 0x6f, 0x75,
	0x6e, 0x63, 0x65, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65,
	0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x65, 0x64, 0x10, 0x04, 0x42, 0x59, 0x5a, 0x57, 0x67, 0x69,
	0x74, 0x6c, 0x61, 0x62, 0x2e, 0x64, 0x69, 0x67, 0x69, 0x75, 0x2e, 0x61, 0x69, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x61, 0x62, 0x6f, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x79, 0x2f, 0x65, 0x79, 0x77, 0x61, 0x2d, 0x70, 0x32, 0x70, 0x2d, 0x62, 0x72, 0x69, 0x64,
	0x67, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x3b, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65,
}

var (
	file_message_proto_rawDescOnce sync.Once
	file_message_proto_rawDescData = file_message_proto_rawDesc
)

func file_message_proto_rawDescGZIP() []byte {
	file_message_proto_rawDescOnce.Do(func() {
		file_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_proto_rawDescData)
	})
	return file_message_proto_rawDescData
}

var file_message_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_message_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_message_proto_goTypes = []interface{}{
	(MsgType)(0),             // 0: message.Msg_type
	(*Message)(nil),          // 1: message.Message
	(*ConsensusRequest)(nil), // 2: message.ConsensusRequest
	(*DrandRequest)(nil),     // 3: message.DrandRequest
	(*UptimeRequest)(nil),    // 4: message.UptimeRequest
}
var file_message_proto_depIdxs = []int32{
	2, // 0: message.Message.consensus:type_name -> message.ConsensusRequest
	3, // 1: message.Message.drand:type_name -> message.DrandRequest
	4, // 2: message.Message.uptime:type_name -> message.UptimeRequest
	0, // 3: message.ConsensusRequest.Msg_type:type_name -> message.Msg_type
	2, // 4: message.ConsensusRequest.history:type_name -> message.ConsensusRequest
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_message_proto_init() }
func file_message_proto_init() {
	if File_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConsensusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DrandRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_message_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UptimeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_message_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Message_Consensus)(nil),
		(*Message_Drand)(nil),
		(*Message_Uptime)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_proto_goTypes,
		DependencyIndexes: file_message_proto_depIdxs,
		EnumInfos:         file_message_proto_enumTypes,
		MessageInfos:      file_message_proto_msgTypes,
	}.Build()
	File_message_proto = out.File
	file_message_proto_rawDesc = nil
	file_message_proto_goTypes = nil
	file_message_proto_depIdxs = nil
}
