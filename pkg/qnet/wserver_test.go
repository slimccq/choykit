// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package qnet

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/codec"
	"github.com/gorilla/websocket"
)

func startClient(t *testing.T, addr, path string) {
	time.Sleep(500 * time.Millisecond)
	wurl := url.URL{Scheme: "ws", Host: addr, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(wurl.String(), nil)
	if err != nil {
		t.Fatal("ws dial:", err)
	}
	defer c.Close()

	var pkt choykit.Packet
	pkt.Flags |= choykit.PacketFlagJSONText
	pkt.Command = 1234
	pkt.Seq = 100
	pkt.Referer = 222
	pkt.Body = "ping"

	data, err := json.Marshal(pkt)
	if err != nil {
		t.Fatalf("marshal %v", err)
	}
	if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write: %v", err)
	}

	var nbytes, msgcnt int
	for i := 0; i < 10; i++ {
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("read %v", err)
		}
		msgcnt += 1
		nbytes += len(msg)
		// fmt.Printf("recv server msg: %s\n", string(msg))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	t.Logf("client recv %d messages, #%d bytes\n", msgcnt, nbytes)
}

func TestWebsocketServer(t *testing.T) {
	var addr = "127.0.0.1:10007"
	var path = "/example"
	var incoming = make(chan *choykit.Packet, 1000)
	var cdec = codec.NewServerCodec()
	server := NewWebsocketServer(addr, path, cdec, incoming, 600)
	server.Go()

	go startClient(t, addr, path)

	var nbytes, msgcnt int
	for {
		select {
		case conn, ok := <-server.BacklogChan():
			if !ok {
				return
			}
			fmt.Printf("connection %s connected\n", conn.RemoteAddr())

		case err := <-server.ErrChan():
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			fmt.Printf("endpoint[%v] %v closed\n", endpoint.NodeID(), endpoint.RemoteAddr())
			return

		case pkt, ok := <-incoming:
			if !ok {
				return
			}
			msgcnt++
			text := pkt.DecodeAsString()
			nbytes += len(text)
			pkt.ReplyAny(pkt.Command, "pong")
			// fmt.Printf("recv client message: %v\n", text))
		}
	}
	t.Logf("server recv %d messages, #%d bytes\n", msgcnt, nbytes)
}
