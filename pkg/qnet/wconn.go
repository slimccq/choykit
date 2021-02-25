// Copyright © 2018-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package qnet

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"sync/atomic"
	"time"

	"devpkg.work/choykit/pkg"
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

func NewWsConn(node choykit.NodeID, conn *websocket.Conn, cdec choykit.Codec, errChan chan error,
	incoming chan<- *choykit.Packet, outsize int, stats *choykit.Stats) *WsConn {
	wsconn := &WsConn{
		conn: conn,
	}
	wsconn.ConnBase.init(node, cdec, incoming, outsize, errChan, stats)
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

func (c *WsConn) SendPacket(pkt *choykit.Packet) error {
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
	c.codec = nil
	c.conn = nil
}

func (c *WsConn) sendPacket(pkt *choykit.Packet, allowBatch bool) error {
	if (pkt.Flags | choykit.PacketFlagJSONText) > 0 {
		return c.sendJSONTextMessage(pkt, allowBatch)
	} else {
		return c.sendBinaryMessage(pkt, allowBatch)
	}
}

func (c *WsConn) sendJSONTextMessage(pkt *choykit.Packet, allowBatch bool) error {
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

func batchAppendMessage(w io.Writer, pkt *choykit.Packet) (int, error) {
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

func (c *WsConn) sendBinaryMessage(pkt *choykit.Packet, allowBatch bool) error {
	var count = 1
	var buf bytes.Buffer
	if err := c.codec.Encode(pkt, &buf); err != nil {
		log.Errorf("WsConn: encode message %d, %v", pkt.Command, err)
		return err
	}
	if allowBatch {
		for n := len(c.outbound); n > 0; n-- {
			pkt := <-c.outbound
			if err := c.codec.Encode(pkt, &buf); err != nil {
				log.Errorf("WsConn: encode message %d, %v", pkt.Command, err)
				break
			}
			count++
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
		var pkt = choykit.MakePacket()
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

func (c *WsConn) ReadPacket(pkt *choykit.Packet) error {
	c.conn.SetReadDeadline(choykit.Now().Add(WSConnReadTimeout))
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
		_, err = c.codec.Decode(bytes.NewReader(data), pkt)
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
