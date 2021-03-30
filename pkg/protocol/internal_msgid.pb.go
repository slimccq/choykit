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

// 内部消息协议
type InternalMsgType int32

const (
	// 内部通信协议 [100 - 1000)
	MSG_INTERNAL_RESERVED              InternalMsgType = 0
	MSG_INTERNAL_REGISTER              InternalMsgType = 104
	MSG_INTERNAL_REGISTER_STATUS       InternalMsgType = 105
	MSG_INTERNAL_INTRODUCE             InternalMsgType = 106
	MSG_INTERNAL_INTRODUCE_STATUS      InternalMsgType = 107
	MSG_INTERNAL_SUBSCRIBE             InternalMsgType = 123
	MSG_INTERNAL_SUBSCRIBE_STATUS      InternalMsgType = 124
	MSG_INTERNAL_KEEP_ALIVE            InternalMsgType = 101
	MSG_INTERNAL_KEEP_ALIVE_STATUS     InternalMsgType = 102
	MSG_INTERNAL_INSTANCE_STATE_NOTIFY InternalMsgType = 103
	MSG_INTERNAL_ROUTE_FORWARD         InternalMsgType = 201
	MSG_INTERNAL_ROUTE_MULTICAST       InternalMsgType = 202
)

var InternalMsgType_name = map[int32]string{
	0:   "MSG_INTERNAL_RESERVED",
	104: "MSG_INTERNAL_REGISTER",
	105: "MSG_INTERNAL_REGISTER_STATUS",
	106: "MSG_INTERNAL_INTRODUCE",
	107: "MSG_INTERNAL_INTRODUCE_STATUS",
	123: "MSG_INTERNAL_SUBSCRIBE",
	124: "MSG_INTERNAL_SUBSCRIBE_STATUS",
	101: "MSG_INTERNAL_KEEP_ALIVE",
	102: "MSG_INTERNAL_KEEP_ALIVE_STATUS",
	103: "MSG_INTERNAL_INSTANCE_STATE_NOTIFY",
	201: "MSG_INTERNAL_ROUTE_FORWARD",
	202: "MSG_INTERNAL_ROUTE_MULTICAST",
}

var InternalMsgType_value = map[string]int32{
	"MSG_INTERNAL_RESERVED":              0,
	"MSG_INTERNAL_REGISTER":              104,
	"MSG_INTERNAL_REGISTER_STATUS":       105,
	"MSG_INTERNAL_INTRODUCE":             106,
	"MSG_INTERNAL_INTRODUCE_STATUS":      107,
	"MSG_INTERNAL_SUBSCRIBE":             123,
	"MSG_INTERNAL_SUBSCRIBE_STATUS":      124,
	"MSG_INTERNAL_KEEP_ALIVE":            101,
	"MSG_INTERNAL_KEEP_ALIVE_STATUS":     102,
	"MSG_INTERNAL_INSTANCE_STATE_NOTIFY": 103,
	"MSG_INTERNAL_ROUTE_FORWARD":         201,
	"MSG_INTERNAL_ROUTE_MULTICAST":       202,
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
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc9, 0xcc, 0x2b, 0x49,
	0x2d, 0xca, 0x4b, 0xcc, 0x89, 0xcf, 0x2d, 0x4e, 0xcf, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x52, 0x22, 0xe9, 0xf9, 0xe9, 0xf9, 0x60, 0x9e, 0x3e,
	0x88, 0x05, 0x91, 0xd7, 0x9a, 0xc2, 0xcc, 0xc5, 0xef, 0x09, 0xd5, 0xe8, 0x5b, 0x9c, 0x1e, 0x52,
	0x59, 0x90, 0x2a, 0x24, 0xc9, 0x25, 0xea, 0x1b, 0xec, 0x1e, 0xef, 0xe9, 0x17, 0xe2, 0x1a, 0xe4,
	0xe7, 0xe8, 0x13, 0x1f, 0xe4, 0x1a, 0xec, 0x1a, 0x14, 0xe6, 0xea, 0x22, 0xc0, 0x80, 0x45, 0xca,
	0xdd, 0x33, 0x38, 0xc4, 0x35, 0x48, 0x20, 0x43, 0x48, 0x81, 0x4b, 0x06, 0xab, 0x54, 0x7c, 0x70,
	0x88, 0x63, 0x48, 0x68, 0xb0, 0x40, 0xa6, 0x90, 0x14, 0x97, 0x18, 0x8a, 0x0a, 0x4f, 0xbf, 0x90,
	0x20, 0x7f, 0x97, 0x50, 0x67, 0x57, 0x81, 0x2c, 0x21, 0x45, 0x2e, 0x59, 0xec, 0x72, 0x30, 0xed,
	0xd9, 0x18, 0xda, 0x83, 0x43, 0x9d, 0x82, 0x9d, 0x83, 0x3c, 0x9d, 0x5c, 0x05, 0xaa, 0x31, 0xb4,
	0xc3, 0xe5, 0x60, 0xda, 0x6b, 0x84, 0xa4, 0xb9, 0xc4, 0x51, 0x94, 0x78, 0xbb, 0xba, 0x06, 0xc4,
	0x3b, 0xfa, 0x78, 0x86, 0xb9, 0x0a, 0xa4, 0x0a, 0x29, 0x71, 0xc9, 0xe1, 0x90, 0x84, 0x19, 0x90,
	0x26, 0xa4, 0xc6, 0xa5, 0x84, 0xe6, 0xc4, 0xe0, 0x10, 0x47, 0x3f, 0xa8, 0x0b, 0x5d, 0xe3, 0xfd,
	0xfc, 0x43, 0x3c, 0xdd, 0x22, 0x05, 0xd2, 0x85, 0xe4, 0xb9, 0xa4, 0x50, 0x03, 0xc2, 0x3f, 0x34,
	0xc4, 0x35, 0xde, 0xcd, 0x3f, 0x28, 0xdc, 0x31, 0xc8, 0x45, 0xe0, 0x24, 0xa3, 0x90, 0x22, 0x7a,
	0x48, 0x81, 0x15, 0xf8, 0x86, 0xfa, 0x84, 0x78, 0x3a, 0x3b, 0x06, 0x87, 0x08, 0x9c, 0x62, 0x74,
	0x12, 0xb9, 0xf0, 0x50, 0x8e, 0xe1, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c,
	0x92, 0x63, 0x9c, 0xf1, 0x58, 0x8e, 0x21, 0x89, 0x0d, 0x1c, 0x67, 0xc6, 0x80, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x19, 0x30, 0x14, 0x63, 0xeb, 0x01, 0x00, 0x00,
}
