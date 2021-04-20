// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package qnet

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/fatchoy"
)

const (
	connectionCount   = 1
	totalMessageCount = 200000
)

func startQPSServer(t *testing.T, address string, ctor, done chan struct{}) {
	var incoming = make(chan *fatchoy.Packet, totalMessageCount)
	var encoder = codec.NewServerProtocolCodec()
	var listener = NewTcpServer(encoder, incoming, totalMessageCount)
	if err := listener.Listen(address); err != nil {
		t.Fatalf("BindTCP: %s %v", address, err)
	}

	ctor <- struct{}{} // server listen OK
	var autoId int32 = 1

	for {
		select {
		case endpoint := <-listener.BacklogChan():
			endpoint.SetNodeID(fatchoy.NodeID(autoId))
			endpoint.Go(true, true)
			autoId++

		case err := <-listener.ErrorChan():
			// handle connection error
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			if !endpoint.IsClosing() {
				endpoint.Close()
			}

		case pkt := <-incoming:
			pkt.ReplyAny(pkt.Command, "pong") //返回pong

		case <-done:
			// handle shutdown
			return
		}
	}
}

func startQPSClient(t *testing.T, address string, msgCount int, respChan chan int) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Dial %s: %v", address, err)
	}

	var encoder = codec.NewServerProtocolCodec()
	var buf bytes.Buffer
	for i := 0; i < msgCount; i++ {
		var pkt = fatchoy.MakePacket()
		pkt.Command = uint32(i)
		pkt.Body = "ping"
		buf.Reset()
		if _, err := encoder.Marshal(&buf, nil, pkt); err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := conn.Write(buf.Bytes()); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	for i := 0; i < msgCount; i++ {
		var resp fatchoy.Packet
		if _, err := encoder.Unmarshal(conn, nil, &resp); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		respChan <- 1
	}
}

func TestQPSBenchmark(t *testing.T) {
	var address = "localhost:10001"
	const eachConnectMsgCount = totalMessageCount / connectionCount

	var ctor = make(chan struct{})
	var done = make(chan struct{})
	go startQPSServer(t, address, ctor, done)
	<-ctor // listener OK

	var respChan = make(chan int, totalMessageCount)
	for i := 0; i < connectionCount; i++ {
		go startQPSClient(t, address, eachConnectMsgCount, respChan)
	}

	fmt.Printf("start benchmark %v\n", time.Now())
	var startTime = time.Now()
	for i := 0; i < totalMessageCount; i++ {
		<-respChan
	}
	var elapsed = time.Now().Sub(startTime)
	fmt.Printf("benchmark finished %v\n", time.Now())
	var qps = float64(totalMessageCount) / (float64(elapsed) / float64(time.Second))
	fmt.Printf("Send %d message with %d clients cost %v, QPS: %f\n", totalMessageCount, connectionCount, elapsed, qps)

	close(done)

	fmt.Printf("Benchmark finished\n")
}
