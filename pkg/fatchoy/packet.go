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
	PacketFlagError      = 0x10
	PacketFlagRpc        = 0x20
	PacketFlagJSONText   = 0x80
	PacketFlagCompressed = 0x01
	PacketFlagEncrypted  = 0x02
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
	Command  uint32          `json:"cmd"`            // 消息ID
	Seq      uint32          `json:"seq"`            // 序列号
	Flag     uint16          `json:"flg,omitempty"`  // 标记位
	Body     interface{}     `json:"body,omitempty"` // 消息内容，number/string/bytes/pb.Message
	Endpoint MessageEndpoint `json:"-"`              // 关联的endpoint
}

func MakePacket() *Packet {
	return &Packet{}
}

func NewPacket(command, seq uint32, flag uint16, body interface{}) *Packet {
	var pkt = &Packet{}
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
	m.Body = nil
	m.Endpoint = nil
}

func (m *Packet) Clone() *Packet {
	var pkt = &Packet{}
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
	m.Body = ec
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
	var pkt = NewPacket(command, m.Seq, m.Flag, data)
	return m.Endpoint.SendPacket(pkt)
}

// 返回一个错误码消息
func (m *Packet) Refuse(command int32, errno uint32) error {
	var pkt = NewPacket(uint32(command), m.Seq, m.Flag|PacketFlagError, errno)
	return m.Endpoint.SendPacket(pkt)
}

func (m Packet) String() string {
	var nodeID NodeID
	if m.Endpoint != nil {
		nodeID = m.Endpoint.NodeID()
	}
	return fmt.Sprintf("%v c:%d seq:%d 0x%x %T", nodeID, m.Command, m.Seq, m.Flag, m.Body)
}
