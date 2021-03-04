// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"devpkg.work/choykit/pkg/codec"
	"net"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	RequestReadTimeout = 60
)

func DialTCP(address string) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ListenTCP(address string) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", addr)
}

// recv一条protobuf消息
func ReadProtoMessage(conn net.Conn, decoder fatchoy.ProtocolDecoder, decrypt fatchoy.MessageEncryptor,
	pkt *fatchoy.Packet, pbMsg proto.Message) error {
	var deadline = fatchoy.Now().Add(time.Duration(RequestReadTimeout) * time.Second)
	conn.SetReadDeadline(deadline)
	_, err := decoder.Unmarshal(conn, pkt)
	if err != nil {
		log.Errorf("decode message %d: %v", pkt.Command, err)
		return err
	}
	if err = codec.DecodePacket(pkt, decrypt); err != nil {
		return err
	}
	if ec := pkt.Errno(); ec > 0 {
		return errors.Errorf("message %d encountered error: %d", pkt.Command, ec)
	}
	if err := pkt.DecodeMsg(pbMsg); err != nil {
		return err
	}
	return nil
}

// send一个packet
func SendPacketMessage(conn net.Conn, encoder fatchoy.ProtocolEncoder, encrypt fatchoy.MessageEncryptor,
	pkt *fatchoy.Packet) error {
	if err := codec.EncodePacket(pkt, 0, encrypt); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := encoder.Marshal(&buf, pkt); err != nil {
		log.Errorf("encode message %d: %v", pkt.Command, err)
		return err
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Errorf("write message %d: %v", pkt.Command, err)
		return nil
	}
	return nil
}

// send一条protobuf消息
func SendProtoMessage(conn net.Conn, encoder fatchoy.ProtocolCodec, encrypt fatchoy.MessageEncryptor,
	command int32, outMsg proto.Message) error {
	var buf bytes.Buffer
	var pkt = fatchoy.MakePacket()
	pkt.Command = uint32(command)
	pkt.Body = outMsg
	if err := codec.EncodePacket(pkt, 0, encrypt); err != nil {
		return err
	}
	if err := encoder.Marshal(&buf, pkt); err != nil {
		log.Errorf("encode message %d: %v", command, err)
		return err
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Errorf("write message %d: %v", command, err)
		return nil
	}
	return nil
}

// send并且立即等待recv
func RequestMessage(conn net.Conn, encoder fatchoy.ProtocolCodec, encrypt fatchoy.MessageEncryptor,
	reqCommand int32, msgReq, msgResp proto.Message) error {
	if err := SendProtoMessage(conn, encoder, encrypt, reqCommand, msgReq); err != nil {
		return err
	}
	var pkt = fatchoy.MakePacket()
	if err := ReadProtoMessage(conn, encoder, encrypt, pkt, msgResp); err != nil {
		return err
	}
	return nil
}
