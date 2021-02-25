// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"net"
	"time"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/log"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	RequestReadTimeout = 15
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

func ReadProtoMessage(conn net.Conn, cdec choykit.Codec, msgOut proto.Message) (*choykit.Packet, error) {
	var pkt = choykit.MakePacket()
	var deadline = choykit.Now().Add(time.Duration(RequestReadTimeout) * time.Second)
	conn.SetReadDeadline(deadline)
	_, err := cdec.Decode(conn, pkt)
	if err != nil {
		log.Errorf("decode message %d: %v", pkt.Command, err)
		return nil, err
	}
	if ec := pkt.Errno(); ec > 0 {
		return nil, errors.Errorf("message %d encountered error: %d", pkt.Command, ec)
	}
	if err := pkt.DecodeMsg(msgOut); err != nil {
		return pkt, err
	}
	return pkt, nil
}

func SendPacketMessage(conn net.Conn, cdec choykit.Codec, pkt *choykit.Packet) error {
	var buf bytes.Buffer
	if err := cdec.Encode(pkt, &buf); err != nil {
		log.Errorf("encode message %d: %v", pkt.Command, err)
		return err
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Errorf("write message %d: %v", pkt.Command, err)
		return nil
	}
	return nil
}

func SendProtoMessage(conn net.Conn, cdec choykit.Codec, command int32, msgIn proto.Message) error {
	var buf bytes.Buffer
	var pkt = choykit.MakePacket()
	pkt.Command = uint32(command)
	pkt.Body = msgIn
	if err := cdec.Encode(pkt, &buf); err != nil {
		log.Errorf("encode message %d: %v", command, err)
		return err
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Errorf("write message %d: %v", command, err)
		return nil
	}
	return nil
}

// send request message and wait for response message
func RequestMessage(conn net.Conn, cdec choykit.Codec, reqCommand int32, msgReq, msgResp proto.Message) error {
	if err := SendProtoMessage(conn, cdec, reqCommand, msgReq); err != nil {
		return err
	}
	if _, err := ReadProtoMessage(conn, cdec, msgResp); err != nil {
		return err
	}
	return nil
}
