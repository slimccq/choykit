// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"errors"
	"fmt"

	"devpkg.work/choykit/pkg/protocol"
	"github.com/gogo/protobuf/proto"
)

const (
	PacketFlagError    = 0x0100
	PacketFlagRefer    = 0x0200
	PacketFlagRpc      = 0x0400
	PacketFlagJSONText = 0x0800
	PacketFlagCompressed = 0x0001
	PacketFlagEncrypted  = 0x0002

	PacketFlagBitsMask = 0xFF00 // 低8位的标志用于传输处理，完成传输后需要清除，不能再返回给ack
)

type PacketHandler func(*Packet) error
type PacketFilter func(*Packet) bool

var (
	ErrOutboundQueueOverflow   = errors.New("outbound queue overflow")
	ErrPacketContextEmpty      = errors.New("packet dispatch context is empty")
	ErrDestinationNotReachable = errors.New("destination not reachable")
)

// 应用层消息
type Packet struct {
	Command  uint32          `json:"cmd"`            // 指令ID
	Seq      uint16          `json:"seq"`            // 序列号
	Flag     uint16          `json:"flg,omitempty"`  // 标记位
	Node     NodeID          `json:"node,omitempty"` // 目标节点
	Body     interface{}     `json:"body,omitempty"` // 消息内容，integer/bytes/string/pb.Message
	Endpoint MessageEndpoint `json:"-"`              // 关联的endpoint
}

func MakePacket() *Packet {
	return &Packet{}
}

func NewPacket(node NodeID, command uint32, flag, seq uint16, body interface{}) *Packet {
	var pkt = &Packet{}
	pkt.Node = node
	pkt.Command = command
	pkt.Flag = flag
	pkt.Seq = seq
	pkt.Body = body
	return pkt
}

func (m *Packet) Reset() {
	m.Command = 0
	m.Seq = 0
	m.Flag = 0
	m.Node = 0
	m.Body = nil
	m.Endpoint = nil
}

func (m *Packet) Clone() *Packet {
	var pkt = &Packet{}
	pkt.Node = m.Node
	pkt.Command = m.Command
	pkt.Flag = m.Flag
	pkt.Seq = m.Seq
	pkt.Body = m.Body
	pkt.Endpoint = m.Endpoint
	return pkt
}

func (m *Packet) Errno() uint32 {
	if (m.Flag & PacketFlagError) != 0 {
		return m.Body.(uint32)
	}
	return 0
}

func (m *Packet) SetErrno(ec uint32) {
	m.Flag |= PacketFlagError
	m.Body = uint32(ec)
}

func (m *Packet) EncodeBody() ([]byte, error) {
	if m.Body == nil {
		return nil, nil
	}
	data, err := EncodeValue(m.Body)
	if err == nil {
		m.Body = nil
	}
	return data, err
}

// 解码为string
func (m *Packet) DecodeBodyAsString() string {
	s := DecodeAsString(m.Body)
	m.Body = nil
	return s
}

// 解码成protobuf message
func (m *Packet) DecodeMsg(msg proto.Message) error {
	err := DecodeAsMsg(m.Body, msg)
	m.Body = nil
	return err
}

func (m *Packet) Run() error {
	if endpoint := m.Endpoint; endpoint != nil {
		if ctx := endpoint.Context(); ctx != nil {
			return ctx.Service().Dispatch(m)
		}
	}
	return ErrPacketContextEmpty
}

// Reply with response message
func (m *Packet) Ack(msgId int32, ack proto.Message) error {
	return m.ReplyAny(uint32(msgId), ack)
}

func (m *Packet) Reply(ack proto.Message) error {
	var mid = protocol.GetMessageIDOf(ack)
	return m.ReplyAny(uint32(mid), ack)
}

func (m *Packet) ReplyAny(command uint32, data interface{}) error {
	var flags = m.Flag & PacketFlagBitsMask
	var pkt = NewPacket(m.Endpoint.NodeID(), command, flags, m.Seq, data)
	return m.Endpoint.SendPacket(pkt)
}

// Refuse with errno
func (m *Packet) Refuse(command int32, errno uint32) error {
	var flags = m.Flag & PacketFlagBitsMask
	var pkt = NewPacket(m.Endpoint.NodeID(), uint32(command), flags|PacketFlagError, m.Seq, uint32(errno))
	return m.Endpoint.SendPacket(pkt)
}

func (m Packet) String() string {
	return fmt.Sprintf("%v c:%d seq:%d 0x%x %T", m.Node, m.Command, m.Seq, m.Flag, m.Body)
}
