// Copyright © 2018-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package qnet

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/codec"
)

//不断发送ping接收pong
func startRawClient(t *testing.T, id int, address string, msgCount int) {
	//t.Logf("client %d start connect %s", id, address)
	conn, err := DialTCP(address)
	if err != nil {
		t.Fatalf("Dial %s: %v", address, err)
	}

	var cdec = codec.NewServerCodec()
	var pkt = choykit.MakePacket()
	for i := 1; i <= msgCount; i++ {
		pkt.Command = uint32(i)
		pkt.Seq = uint16(i)
		pkt.Body = "ping"
		var buf bytes.Buffer
		if err := cdec.Encode(pkt, &buf); err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := conn.Write(buf.Bytes()); err != nil {
			t.Fatalf("Write: %v", err)
		}
		var resp choykit.Packet
		if _, err := cdec.Decode(conn, &resp); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if resp.Seq != pkt.Seq {
			t.Fatalf("session mismatch, %d != %d", resp.Seq, pkt.Seq)
		}
		payload, _ := resp.Encode()
		if v := string(payload); v != "pong" {
			t.Fatalf("invalid response message: %s", v)
		}
		//fmt.Printf("message %d OK\n", i)
	}
	//fmt.Printf("Connection %v OK\n", conn.RemoteAddr())
}

func startMyListener(t *testing.T, address string, sig, done chan struct{}) {
	var cdec = codec.NewServerCodec()
	var incoming = make(chan *choykit.Packet, 100)
	var server = NewTcpServer(cdec, incoming, 60)
	if err := server.Listen(address); err != nil {
		t.Fatalf("BindTCP: %s %v", address, err)
	}

	sig <- struct{}{} // server listen OK

	var autoId uint32
	var recvNum = 0
	var t2 = time.NewTimer(time.Minute) // this case should pass within 1 minute
	const totalMsgNum = maxPingpong * maxConnection

	for {
		select {
		case endpoint := <-server.BacklogChan():
			// handle new connection
			// var addr = endpoint.RemoteAddr()
			// fmt.Printf("endpoint %v connected\n", addr)
			autoId++
			endpoint.SetNodeID(choykit.NodeID(autoId))
			endpoint.Go(true, true)

		case err := <-server.ErrorChan():
			// handle connection error
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			// fmt.Printf("endpoint[%v] %v closed\n", endpoint.Node(), endpoint.RemoteAddr())
			if !endpoint.IsClosing() {
				endpoint.Close()
			}

		case pkt := <-incoming:
			pkt.ReplyAny(pkt.Command, "pong") //返回pong

			// all message recv, close server
			recvNum++
			if recvNum > 0 && recvNum%100 == 0 {
				//fmt.Printf("recv messages: %d/%d\n", recvNum, totalMsgNum)
			}
			if recvNum == totalMsgNum {
				fmt.Printf("all messages recv OK, shutdown\n")
				go func() { close(done) }()
			}

		case <-t2.C:
			fmt.Printf("timeout to end\n")
			t.FailNow()
			return

		case <-done:
			// handle shutdown
			fmt.Printf("listener done\n")
			return
		}
	}
}

func TestExampleServerUsage(t *testing.T) {
	var testTcpAddress = "localhost:10004"
	var listenOK = make(chan struct{})
	var done = make(chan struct{})
	go startMyListener(t, testTcpAddress, listenOK, done)

	<-listenOK // wait listen init done
	t.Logf("server listen OK")

	// start client connections
	for i := 0; i < maxConnection; i++ {
		time.Sleep(10 * time.Millisecond)
		go startRawClient(t, i+1, testTcpAddress, maxPingpong)
	}

	<-done // wait till done
}
