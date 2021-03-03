// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: internal_msgid.proto

package protocol

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
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

// 内部通信协议 [100 - 1000)
type InternalMsgType int32

const (
	MSG_INTERNAL_RESERVED              InternalMsgType = 0
	MSG_INTERNAL_KEEP_ALIVE            InternalMsgType = 101
	MSG_INTERNAL_KEEP_ALIVE_STATUS     InternalMsgType = 102
	MSG_INTERNAL_INSTANCE_STATE_NOTIFY InternalMsgType = 103
	MSG_INTERNAL_REGISTER              InternalMsgType = 104
	MSG_INTERNAL_REGISTER_STATUS       InternalMsgType = 105
	MSG_INTERNAL_INTRODUCE             InternalMsgType = 106
	MSG_INTERNAL_INTRODUCE_STATUS      InternalMsgType = 107
	MSG_INTERNAL_SUBSCRIBE             InternalMsgType = 123
	MSG_INTERNAL_SUBSCRIBE_STATUS      InternalMsgType = 124
)

var InternalMsgType_name = map[int32]string{
	0:   "MSG_INTERNAL_RESERVED",
	101: "MSG_INTERNAL_KEEP_ALIVE",
	102: "MSG_INTERNAL_KEEP_ALIVE_STATUS",
	103: "MSG_INTERNAL_INSTANCE_STATE_NOTIFY",
	104: "MSG_INTERNAL_REGISTER",
	105: "MSG_INTERNAL_REGISTER_STATUS",
	106: "MSG_INTERNAL_INTRODUCE",
	107: "MSG_INTERNAL_INTRODUCE_STATUS",
	123: "MSG_INTERNAL_SUBSCRIBE",
	124: "MSG_INTERNAL_SUBSCRIBE_STATUS",
}

var InternalMsgType_value = map[string]int32{
	"MSG_INTERNAL_RESERVED":              0,
	"MSG_INTERNAL_KEEP_ALIVE":            101,
	"MSG_INTERNAL_KEEP_ALIVE_STATUS":     102,
	"MSG_INTERNAL_INSTANCE_STATE_NOTIFY": 103,
	"MSG_INTERNAL_REGISTER":              104,
	"MSG_INTERNAL_REGISTER_STATUS":       105,
	"MSG_INTERNAL_INTRODUCE":             106,
	"MSG_INTERNAL_INTRODUCE_STATUS":      107,
	"MSG_INTERNAL_SUBSCRIBE":             123,
	"MSG_INTERNAL_SUBSCRIBE_STATUS":      124,
}

func (x InternalMsgType) String() string {
	return proto.EnumName(InternalMsgType_name, int32(x))
}

func (InternalMsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_e73ab34aa251ccc7, []int{0}
}

func init() {
	proto.RegisterEnum("protocol.InternalMsgType", InternalMsgType_name, InternalMsgType_value)
}

func init() { proto.RegisterFile("internal_msgid.proto", fileDescriptor_e73ab34aa251ccc7) }

var fileDescriptor_e73ab34aa251ccc7 = []byte{
	// 258 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc9, 0xcc, 0x2b, 0x49,
	0x2d, 0xca, 0x4b, 0xcc, 0x89, 0xcf, 0x2d, 0x4e, 0xcf, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x52, 0x22, 0xe9, 0xf9, 0xe9, 0xf9, 0x60, 0x9e, 0x3e,
	0x88, 0x05, 0x91, 0xd7, 0xba, 0xc0, 0xc4, 0xc5, 0xef, 0x09, 0xd5, 0xe8, 0x5b, 0x9c, 0x1e, 0x52,
	0x59, 0x90, 0x2a, 0x24, 0xc9, 0x25, 0xea, 0x1b, 0xec, 0x1e, 0xef, 0xe9, 0x17, 0xe2, 0x1a, 0xe4,
	0xe7, 0xe8, 0x13, 0x1f, 0xe4, 0x1a, 0xec, 0x1a, 0x14, 0xe6, 0xea, 0x22, 0xc0, 0x20, 0x24, 0xcd,
	0x25, 0x8e, 0x22, 0xe5, 0xed, 0xea, 0x1a, 0x10, 0xef, 0xe8, 0xe3, 0x19, 0xe6, 0x2a, 0x90, 0x2a,
	0xa4, 0xc4, 0x25, 0x87, 0x43, 0x32, 0x3e, 0x38, 0xc4, 0x31, 0x24, 0x34, 0x58, 0x20, 0x4d, 0x48,
	0x8d, 0x4b, 0x09, 0x45, 0x8d, 0xa7, 0x5f, 0x70, 0x88, 0xa3, 0x9f, 0x33, 0x44, 0x85, 0x6b, 0xbc,
	0x9f, 0x7f, 0x88, 0xa7, 0x5b, 0xa4, 0x40, 0x3a, 0x16, 0x37, 0xb8, 0x7b, 0x06, 0x87, 0xb8, 0x06,
	0x09, 0x64, 0x08, 0x29, 0x70, 0xc9, 0x60, 0x95, 0x82, 0x59, 0x92, 0x29, 0x24, 0xc5, 0x25, 0x86,
	0x66, 0x49, 0x48, 0x90, 0xbf, 0x4b, 0xa8, 0xb3, 0xab, 0x40, 0x96, 0x90, 0x22, 0x97, 0x2c, 0x76,
	0x39, 0x98, 0xf6, 0x6c, 0x0c, 0xed, 0xc1, 0xa1, 0x4e, 0xc1, 0xce, 0x41, 0x9e, 0x4e, 0xae, 0x02,
	0xd5, 0x18, 0xda, 0xe1, 0x72, 0x30, 0xed, 0x35, 0x4e, 0x22, 0x17, 0x1e, 0xca, 0x31, 0x9c, 0x78,
	0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x33, 0x1e, 0xcb, 0x31, 0x24,
	0xb1, 0x81, 0xc3, 0xdb, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x32, 0x8d, 0x03, 0xf7, 0xa7, 0x01,
	0x00, 0x00,
}