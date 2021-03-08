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

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"github.com/pkg/errors"
)

var (
	TConnReadTimeout = 100
)

// TCP connection
type TcpConn struct {
	ConnBase
	conn   net.Conn  // TCP connection object
	reader io.Reader // bufio reader
}

func NewTcpConn(node fatchoy.NodeID, conn net.Conn, encoder fatchoy.ProtocolCodec, errChan chan error,
	incoming chan<- *fatchoy.Packet, outsize int32, stats *fatchoy.Stats) *TcpConn {
	tconn := &TcpConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	tconn.ConnBase.init(node, encoder, incoming, outsize, errChan, stats)
	tconn.addr = conn.RemoteAddr().String()
	return tconn
}

func (t *TcpConn) RawConn() net.Conn {
	return t.conn
}

func (t *TcpConn) OutboundQueue() chan *fatchoy.Packet {
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

func (t *TcpConn) SendPacket(pkt *fatchoy.Packet) error {
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
	t.finally(ErrConnForceClose) // 阻塞等待投递剩余的消息
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
	go t.finally(err) // 不阻塞等待
}

func (t *TcpConn) finally(err error) {
	close(t.done)
	t.wg.Wait()
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseWrite()
	} else {
		t.conn.Close()
	}
	// 把error投递给监听的channel
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
	t.encoder = nil
	t.conn = nil
	t.reader = nil
}

func (t *TcpConn) flush() {
	var buf bytes.Buffer
	n := len(t.outbound)
	for i := 0; i < n; i++ {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				break
			}
			if err := t.encoder.Marshal(&buf, pkt); err != nil {
				log.Errorf("TcpConn: encode message %d: %v", pkt.Command, err)
				return
			}
		default:
			break
		}
	}
	nbytes, err := t.conn.Write(buf.Bytes())
	if err != nil {
		log.Errorf("TcpConn: node %v write message: %v", t.node, err)
		return
	}
	t.stats.Add(StatPacketsSent, int64(n))
	t.stats.Add(StatBytesSent, int64(nbytes))
}

func (t *TcpConn) writePump() {
	defer func() {
		t.flush()
		t.wg.Done()
		log.Debugf("TcpConn: node %v writer stopped", t.node)
	}()

	log.Debugf("TcpConn: node %v(%v) writer started", t.node, t.addr)
	for {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				return
			}
			var buf bytes.Buffer
			var err error
			if err = t.encoder.Marshal(&buf, pkt); err != nil {
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

func (t *TcpConn) readPacket() (*fatchoy.Packet, error) {
	deadline := fatchoy.Now().Add(time.Duration(TConnReadTimeout) * time.Second)
	t.conn.SetReadDeadline(deadline)
	var pkt = fatchoy.MakePacket()
	nbytes, err := t.encoder.Unmarshal(t.reader, pkt)
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

func (t *TcpConn) testShouldExit() bool {
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
	log.Debugf("TcpConn: node %v(%v) reader started", t.node, t.addr)
	for {
		pkt, err := t.readPacket()
		if err != nil {
			t.ForceClose(err) // I/O超时或者发生错误，强制关闭连接
			return
		}
		t.inbound <- pkt // 如果channel满了，这里会阻塞

		// test if we should exit
		if t.testShouldExit() {
			return
		}
	}
}
