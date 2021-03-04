// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"sync/atomic"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	WSCONN_MAX_PAYLOAD = 16 * 1024 // 16k
)

var (
	WSConnReadTimeout = 100 * time.Second
)

// Websocket connection
type WsConn struct {
	ConnBase
	conn *websocket.Conn // websocket conn
}

func NewWsConn(node fatchoy.NodeID, conn *websocket.Conn, encoder fatchoy.ProtocolCodec, errChan chan error,
	incoming chan<- *fatchoy.Packet, outsize int, stats *fatchoy.Stats) *WsConn {
	wsconn := &WsConn{
		conn: conn,
	}
	wsconn.ConnBase.init(node, encoder, incoming, outsize, errChan, stats)
	wsconn.addr = conn.RemoteAddr().String()
	conn.SetReadLimit(WSCONN_MAX_PAYLOAD)
	conn.SetPingHandler(wsconn.handlePing)
	return wsconn
}

func (c *WsConn) RawConn() net.Conn {
	return c.conn.UnderlyingConn()
}

func (c *WsConn) Go(writer, reader bool) {
	if writer {
		c.wg.Add(1)
		go c.writePump()
	}
}

func (c *WsConn) SendPacket(pkt *fatchoy.Packet) error {
	if c.IsClosing() {
		return ErrConnIsClosing
	}
	select {
	case c.outbound <- pkt:
		return nil
	default:
		log.Errorf("message %d ignored due to queue overflow", pkt.Command)
		return ErrConnOutboundOverflow
	}
}

func (c *WsConn) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		// log.Errorf("WsConn: connection %v is already closed", c.node)
		return nil
	}
	c.finally(ErrConnForceClose)
	return nil
}

func (c *WsConn) ForceClose(err error) {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		// log.Errorf("WsConn: connection %v is already closed", c.node)
		return
	}
	go c.finally(err)
}

func (c *WsConn) finally(err error) {
	close(c.done)
	c.wg.Wait()
	if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		log.Errorf("WsConn: write close message, %v", err)
	}
	if err := c.conn.Close(); err != nil {
		log.Errorf("WsConn: close connection %v, %v", c.node, err)
	}
	if c.errChan != nil {
		select {
		case c.errChan <- NewError(err, c):
		default:
		}
	}
	close(c.outbound)
	c.outbound = nil
	c.inbound = nil
	c.encoder = nil
	c.conn = nil
}

func (c *WsConn) sendPacket(pkt *fatchoy.Packet, allowBatch bool) error {
	if (pkt.Flag | fatchoy.PacketFlagJSONText) > 0 {
		return c.sendJSONTextMessage(pkt, allowBatch)
	} else {
		return c.sendBinaryMessage(pkt, allowBatch)
	}
}

func (c *WsConn) sendJSONTextMessage(pkt *fatchoy.Packet, allowBatch bool) error {
	if !allowBatch {
		data, err := json.Marshal(pkt)
		if err != nil {
			log.Errorf("WsConn: JSON marshal message %d, %v", pkt.Command, err)
			return err
		}
		if err = c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Errorf("WsConn: send text message %d, %v", pkt.Command, err)
			return err
		}
		c.stats.Add(StatPacketsSent, int64(1))
		c.stats.Add(StatBytesSent, int64(len(data)))
		return nil
	}
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Errorf("WsConn: NextWriter for %d, %v", pkt.Command, err)
		return err
	}
	var count = 1
	nbytes, err := batchAppendMessage(w, pkt)
	if err != nil {
		return err
	}
	for n := len(c.outbound); n > 0; n-- {
		pkt := <-c.outbound
		if sz, err := batchAppendMessage(w, pkt); err != nil {
			break
		} else {
			nbytes += sz
		}
		count++
	}
	if err := w.Close(); err != nil {
		log.Errorf("WsConn: send text message %d, %v", pkt.Command, err)
		return err
	}
	c.stats.Add(StatPacketsSent, int64(count))
	c.stats.Add(StatBytesSent, int64(nbytes))

	return nil
}

func batchAppendMessage(w io.Writer, pkt *fatchoy.Packet) (int, error) {
	data, err := json.Marshal(pkt)
	if err != nil {
		log.Errorf("WsConn: JSON marshal message %d, %v", pkt.Command, err)
		return 0, err
	}
	_, err = w.Write(data)
	if err != nil {
		log.Errorf("WsConn: append text message %d, %v", pkt.Command, err)
		return 0, err
	}
	return len(data), nil
}

func (c *WsConn) sendBinaryMessage(pkt *fatchoy.Packet, allowBatch bool) error {
	var count = 1
	if allowBatch {
		count = len(c.outbound) // 发送队列里的所有消息
	}
	var buf bytes.Buffer
	for i := 0; i < count; i++ {
		select {
		case pkt := <-c.outbound:
			if err := c.encoder.Marshal(&buf, pkt); err != nil {
				log.Errorf("WsConn: encode message %d, %v", pkt.Command, err)
				return err
			}
		default:
			break
		}
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		log.Errorf("WsConn: send message %d, %v", pkt.Command, err)
		return err
	}
	c.stats.Add(StatPacketsSent, int64(count))
	c.stats.Add(StatBytesSent, int64(buf.Len()))
	return nil
}

func (c *WsConn) writePump() {
	defer c.wg.Done()
	defer log.Debugf("node %v writer exit", c.node)
	log.Debugf("node %v writer started at %v", c.node, c.addr)
	for {
		select {
		case pkt, ok := <-c.outbound:
			if !ok {
				return
			}
			c.sendPacket(pkt, true)

		case <-c.done:
			return
		}
	}
}

func (c *WsConn) readLoop() {
	for {
		var pkt = fatchoy.MakePacket()
		if err := c.ReadPacket(pkt); err != nil {
			log.Errorf("read message: %v", err)
			break
		}
		pkt.Endpoint = c
		c.inbound <- pkt

		// check if we should exit
		select {
		case <-c.done:
			return
		default:
		}
	}
}

func (c *WsConn) ReadPacket(pkt *fatchoy.Packet) error {
	c.conn.SetReadDeadline(fatchoy.Now().Add(WSConnReadTimeout))
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	c.stats.Add(StatPacketsSent, int64(1))
	c.stats.Add(StatBytesSent, int64(len(data)))

	switch msgType {
	case websocket.TextMessage:
		// log.Debugf("recv message: %s", data)
		return json.Unmarshal(data, pkt)

	case websocket.BinaryMessage:
		_, err = c.encoder.Unmarshal(bytes.NewReader(data), pkt)
		return err

	case websocket.PingMessage, websocket.PongMessage:

	default:
		return errors.Errorf("unexpected websock message type %d", msgType)
	}
	return nil
}

func (c *WsConn) handlePing(data string) error {
	log.Infof("ping message: %s", data)
	return nil
}
