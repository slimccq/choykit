// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package qnet

import (
	"fmt"
	"net"
	"testing"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/x/strutil"
)

const (
	maxConnection = 100
	maxPingpong   = 1000
)

func init() {
	fatchoy.StartClock()
}

func handleConn(conn net.Conn, encoder fatchoy.ProtocolCodec) {
	var count = 0
	//file, _ := conn.File()
	tconn := NewTcpConn(0, conn, encoder, nil, nil, 1000, nil)
	tconn.Go(true, false)
	defer tconn.Close()
	for {
		conn.SetReadDeadline(time.Now().Add(time.Minute))
		var pkt = fatchoy.MakePacket()
		if _, err := encoder.Unmarshal(conn, nil, pkt); err != nil {
			fmt.Printf("Decode: %v\n", err)
			break
		}

		// fmt.Printf("%d srecv: %s\n", file.Fd(), pkt.Body)
		pkt.Body = fmt.Sprintf("pong %d", pkt.Command)
		tconn.SendPacket(pkt)
		//fmt.Printf("message %d OK\n", count)
		count++
		if count == maxPingpong {
			break
		}
	}
	stats := tconn.Stats()
	fmt.Printf("sent %d packets, %s\n", stats.Get(StatPacketsSent), strutil.PrettyBytes(stats.Get(StatBytesSent)))
}

func startMyServer(t *testing.T, ln net.Listener, encoder fatchoy.ProtocolCodec) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			//t.Logf("Listener: Accept %v", err)
			return
		}
		go handleConn(conn, encoder)
	}
}

func tconnReadLoop(errchan chan error, inbound chan *fatchoy.Packet) {
	for {
		select {
		case pkt, ok := <-inbound:
			if !ok {
				return
			}
			pkt.Command += 1
			pkt.ReplyAny(pkt.Command, fmt.Sprintf("ping %d", pkt.Command))

		case <-errchan:
			return
		}
	}
}

func TestExampleTcpConn(t *testing.T) {
	TConnReadTimeout = 30

	var testTcpAddress = "localhost:10002"

	ln, err := net.Listen("tcp", testTcpAddress)
	if err != nil {
		t.Fatalf("Listen %v", err)
	}
	defer ln.Close()

	go startMyServer(t, ln, codec.ServerProtocolCodec)

	conn, err := net.Dial("tcp", testTcpAddress)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	//file, _ := conn.File()
	inbound := make(chan *fatchoy.Packet, 1000)
	errchan := make(chan error, 4)
	tconn := NewTcpConn(0, conn, codec.ServerProtocolCodec, errchan, inbound, 1000, nil)
	tconn.SetNodeID(fatchoy.NodeID(0x12345))
	tconn.Go(true, true)
	defer tconn.Close()
	stats := tconn.Stats()
	var pkt = fatchoy.MakePacket()
	pkt.Command = 1
	pkt.Body = "ping"
	tconn.SendPacket(pkt)
	tconnReadLoop(errchan, inbound)
	fmt.Printf("recv %d packets, %s\n", stats.Get(StatPacketsRecv), strutil.PrettyBytes(stats.Get(StatBytesRecv)))
}
