// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"bufio"
	"devpkg.work/choykit/pkg/codec"
	"io"
	"net"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/qnet"
	"devpkg.work/choykit/pkg/protocol"
)

//
func (g *Service) createClientListener(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Errorf("cannot listen to this interface[%s]: %v", addr, err)
		return err
	}
	// listen to all interface
	caddr := ":" + port
	ln, err := qnet.ListenTCP(caddr)
	if err != nil {
		log.Errorf("listen client interface %s: %v", addr, err)
		return err
	}
	log.Infof("listen client at %s", addr)
	g.cListeners = append(g.cListeners, ln)
	return nil
}

// 侦听客户端连接
func (g *Service) serveClientSession(ln net.Listener) {
	addr := ln.Addr()
	log.Infof("serve client at %v", addr)
	defer log.Infof("stop serving client at %v", addr)
	defer g.wg.Done()
	for !g.IsClosing() {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Warnf("temporary accept error: %v", ne)
				time.Sleep(time.Millisecond * 200)
				continue
			}
			log.Errorf("accept error: %v", err)
			return
		}
		var node = g.nextSession()
		if count := g.sessions.Size() + 1; count > int(g.pcu) {
			g.pcu = uint32(count) // dirty write is OK
		}
		go g.handleClientConn(conn, node)
	}
}

func (g *Service) handleClientConn(conn net.Conn, node fatchoy.NodeID) {
	var sid = NodeToSession(node)
	log.Infof("TCP client #%d connected, %v", sid, conn.RemoteAddr())
	session, err := g.handshakeClient(conn, node)
	if err != nil {
		conn.Close()
		return
	}
	defer log.Infof("client #%d(%v) disconnected", sid, session.RemoteAddr())
	defer g.closeSession(session)
	session.SetContext(g.Context())
	session.Go(true, false)

	var reader = bufio.NewReader(conn)
	var env = g.Context().Env()
	var interval = time.Duration(env.NetSessionReadTimeout) * time.Second
	var encoder = codec.ClientProtocolCodec
	for !g.IsClosing() {
		var pkt = fatchoy.MakePacket()
		conn.SetReadDeadline(fatchoy.Now().Add(interval))
		if _, err := encoder.Unmarshal(reader, pkt); err != nil {
			if err != io.EOF {
				log.Errorf("session %v read packet %v: %v", sid, pkt, err)
			}
			break
		}
		// 限制client消息
		if pkt.Command < uint32(protocol.MSG_CLIENT_START_ID) || pkt.Command > uint32(protocol.MSG_CLIENT_END_ID){
			log.Errorf("session %v illegal packet [%v]", sid, pkt)
			break
		}
		pkt.Endpoint = session
		g.dispatchPacket(pkt)
	}
}

// client握手，client先发一条handshake消息，再发送一条login消息
func (g *Service) handshakeClient(conn net.Conn, node fatchoy.NodeID) (fatchoy.Endpoint, error) {
	// encryption key exchange
	// TODO: net encryption

	// handle login & authentication
	var encoder = codec.ClientProtocolCodec
	var session = qnet.NewTcpConn(node, conn, encoder, nil, g.Context().InboundQueue(),
		g.Context().Env().EndpointOutboundQueueSize, g.cStats)
	if err := g.forwardClientLogin(session); err != nil {
		log.Errorf("handle login error: %v", err)
		return nil, err
	}
	session.SetUserData(NewSessionUserData())
	return session, nil
}
