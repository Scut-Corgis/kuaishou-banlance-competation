// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.26.1
// source: B.proto

package apiB

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

type Sha256Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShaReq string `protobuf:"bytes,1,opt,name=shaReq,proto3" json:"shaReq,omitempty"`
	Num    int32  `protobuf:"varint,2,opt,name=num,proto3" json:"num,omitempty"`
}

func (x *Sha256Request) Reset() {
	*x = Sha256Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_B_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sha256Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sha256Request) ProtoMessage() {}

func (x *Sha256Request) ProtoReflect() protoreflect.Message {
	mi := &file_B_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sha256Request.ProtoReflect.Descriptor instead.
func (*Sha256Request) Descriptor() ([]byte, []int) {
	return file_B_proto_rawDescGZIP(), []int{0}
}

func (x *Sha256Request) GetShaReq() string {
	if x != nil {
		return x.ShaReq
	}
	return ""
}

func (x *Sha256Request) GetNum() int32 {
	if x != nil {
		return x.Num
	}
	return 0
}

type Sha256Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShaResp string `protobuf:"bytes,1,opt,name=shaResp,proto3" json:"shaResp,omitempty"`
}

func (x *Sha256Reply) Reset() {
	*x = Sha256Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_B_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sha256Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sha256Reply) ProtoMessage() {}

func (x *Sha256Reply) ProtoReflect() protoreflect.Message {
	mi := &file_B_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sha256Reply.ProtoReflect.Descriptor instead.
func (*Sha256Reply) Descriptor() ([]byte, []int) {
	return file_B_proto_rawDescGZIP(), []int{1}
}

func (x *Sha256Reply) GetShaResp() string {
	if x != nil {
		return x.ShaResp
	}
	return ""
}

var File_B_proto protoreflect.FileDescriptor

var file_B_proto_rawDesc = []byte{
	0x0a, 0x07, 0x42, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x61, 0x70, 0x69, 0x42, 0x22,
	0x39, 0x0a, 0x0d, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x68, 0x61, 0x52, 0x65, 0x71, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x68, 0x61, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x6e, 0x75, 0x6d, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6e, 0x75, 0x6d, 0x22, 0x27, 0x0a, 0x0b, 0x53, 0x68,
	0x61, 0x32, 0x35, 0x36, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x68, 0x61,
	0x52, 0x65, 0x73, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x68, 0x61, 0x52,
	0x65, 0x73, 0x70, 0x32, 0x3e, 0x0a, 0x01, 0x42, 0x12, 0x39, 0x0a, 0x0d, 0x43, 0x6f, 0x6d, 0x70,
	0x75, 0x74, 0x65, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12, 0x13, 0x2e, 0x61, 0x70, 0x69, 0x42,
	0x2e, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11,
	0x2e, 0x61, 0x70, 0x69, 0x42, 0x2e, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x42, 0x08, 0x5a, 0x06, 0x2e, 0x3b, 0x61, 0x70, 0x69, 0x42, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_B_proto_rawDescOnce sync.Once
	file_B_proto_rawDescData = file_B_proto_rawDesc
)

func file_B_proto_rawDescGZIP() []byte {
	file_B_proto_rawDescOnce.Do(func() {
		file_B_proto_rawDescData = protoimpl.X.CompressGZIP(file_B_proto_rawDescData)
	})
	return file_B_proto_rawDescData
}

var file_B_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_B_proto_goTypes = []interface{}{
	(*Sha256Request)(nil), // 0: apiB.Sha256Request
	(*Sha256Reply)(nil),   // 1: apiB.Sha256Reply
}
var file_B_proto_depIdxs = []int32{
	0, // 0: apiB.B.ComputeSha256:input_type -> apiB.Sha256Request
	1, // 1: apiB.B.ComputeSha256:output_type -> apiB.Sha256Reply
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_B_proto_init() }
func file_B_proto_init() {
	if File_B_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_B_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sha256Request); i {
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
		file_B_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sha256Reply); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_B_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_B_proto_goTypes,
		DependencyIndexes: file_B_proto_depIdxs,
		MessageInfos:      file_B_proto_msgTypes,
	}.Build()
	File_B_proto = out.File
	file_B_proto_rawDesc = nil
	file_B_proto_goTypes = nil
	file_B_proto_depIdxs = nil
}
