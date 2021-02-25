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
	"bufio"
	"bytes"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/log"
)

var (
	TConnReadTimeout = 60
)

// TCP connection
type TcpConn struct {
	ConnBase
	conn   net.Conn  // TCP connection object
	reader io.Reader // bufio reader
}

func NewTcpConn(node choykit.NodeID, conn net.Conn, cdec choykit.Codec, errChan chan error,
	incoming chan<- *choykit.Packet, outsize int, stats *choykit.Stats) *TcpConn {
	tconn := &TcpConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	tconn.ConnBase.init(node, cdec, incoming, outsize, errChan, stats)
	tconn.addr = conn.RemoteAddr().String()
	return tconn
}

func (t *TcpConn) RawConn() net.Conn {
	return t.conn
}

func (t *TcpConn) OutboundQueue() chan *choykit.Packet {
	return t.outbound
}

func (t *TcpConn) Go(writer, reader bool) {
	if writer {
		t.wg.Add(1)
		go t.writePump()
	}
	if reader {
		t.wg.Add(1)
		go t.readPump()
	}
}

func (t *TcpConn) SendPacket(pkt *choykit.Packet) error {
	if t.IsClosing() {
		return ErrConnIsClosing
	}
	select {
	case t.outbound <- pkt:
		return nil
	default:
		log.Errorf("TcpConn: message %d ignored due to queue overflow", pkt.Command)
		return errors.WithStack(ErrConnOutboundOverflow)
	}
}

func (t *TcpConn) Close() error {
	if !atomic.CompareAndSwapInt32(&t.closing, 0, 1) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return nil
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	t.finally(ErrConnForceClose)
	return nil
}

func (t *TcpConn) ForceClose(err error) {
	if !atomic.CompareAndSwapInt32(&t.closing, 0, 1) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	go t.finally(err)
}

func (t *TcpConn) finally(err error) {
	close(t.done)
	t.wg.Wait()
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseWrite()
	} else {
		t.conn.Close()
	}
	if t.errChan != nil {
		select {
		case t.errChan <- NewError(err, t):
		default:
		}
	}
	close(t.outbound)
	t.outbound = nil
	t.inbound = nil
	t.errChan = nil
	t.codec = nil
	t.conn = nil
	t.reader = nil
}

func (t *TcpConn) flush() {
	select {
	case pkt, ok := <-t.outbound:
		if !ok {
			return
		}
		var buf bytes.Buffer
		if err := t.codec.Encode(pkt, &buf); err != nil {
			log.Errorf("TcpConn: encode message %d: %v", pkt.Command, err)
			break
		}
		cnt := 0
		remain := len(t.outbound) // 后续并发写入sendqueue的packet将不会被投递
		for i := 0; i < remain; i++ {
			pkt = <-t.outbound
			cnt++
			if err := t.codec.Encode(pkt, &buf); err != nil {
				log.Errorf("TcpConn: encode batch %d messages %d: %v", cnt, pkt.Command, err)
				break
			}
		}
		nbytes, err := buf.WriteTo(t.conn)
		if err != nil {
			log.Errorf("TcpConn: node %v write message: %v", t.node, err)
		} else {
			t.stats.Add(StatPacketsSent, int64(cnt))
			t.stats.Add(StatBytesSent, nbytes)
		}
		return

	default:
		return
	}
}

func (t *TcpConn) writePump() {
	defer t.wg.Done()
	defer t.flush()
	defer log.Debugf("TcpConn: node %v writer stopped", t.node)
	log.Debugf("TcpConn: node %v writer started at %v", t.node, t.addr)
	for {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				return
			}
			var buf bytes.Buffer
			var err error
			if err = t.codec.Encode(pkt, &buf); err != nil {
				log.Errorf("encode message %v: %v", pkt.Command, err)
				continue
			}
			if n, err := t.conn.Write(buf.Bytes()); err != nil {
				log.Errorf("write message %d to node %v: %v", pkt.Command, t.node, err)
			} else {
				t.stats.Add(StatPacketsSent, 1)
				t.stats.Add(StatBytesSent, int64(n))
			}

		case <-t.done:
			return
		}
	}
}

func (t *TcpConn) readPacket() (*choykit.Packet, error) {
	deadline := choykit.Now().Add(time.Duration(TConnReadTimeout) * time.Second)
	t.conn.SetReadDeadline(deadline)
	var pkt = choykit.MakePacket()
	nbytes, err := t.codec.Decode(t.reader, pkt)
	if err != nil {
		if err != io.EOF {
			log.Errorf("read message from node %v: %v", t.node, err)
		}
		return nil, err
	}
	t.stats.Add(StatPacketsRecv, 1)
	t.stats.Add(StatBytesRecv, int64(nbytes))
	pkt.Endpoint = t
	return pkt, nil
}

func (t *TcpConn) checkIfExit() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

func (t *TcpConn) readPump() {
	defer t.wg.Done()
	defer log.Debugf("TcpConn: node %v reader stopped", t.node)
	log.Debugf("TcpConn: node %v reader started at %v", t.node, t.addr)
	for {
		pkt, err := t.readPacket()
		if err != nil {
			t.ForceClose(err) // I/O超时或者发生错误，强制关闭连接
			return
		}
		t.inbound <- pkt // 如果channel满了，这里会阻塞

		// test if we should exit
		if t.checkIfExit() {
			return
		}
	}
}
