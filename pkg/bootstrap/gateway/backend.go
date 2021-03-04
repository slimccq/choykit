// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"bufio"
	"io"
	"net"
	"time"

	"devpkg.work/choykit/pkg/cluster"
	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"devpkg.work/choykit/pkg/qnet"
)

func (g *Service) createBackendListener(addr string) error {
	ln, err := qnet.ListenTCP(addr)
	if err != nil {
		log.Errorf("listen backend %s: %v", addr, err)
		return err
	}
	g.sListener = ln
	return nil
}

// 侦听backend连接
func (g *Service) serveBackend() {
	addr := g.sListener.Addr()
	log.Infof("serve backend at %v", addr)
	defer log.Infof("stop serving backend at %v", addr)
	defer g.wg.Done()
	for !g.IsClosing() {
		conn, err := g.sListener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Warnf("temporary accept error: %v", ne)
				time.Sleep(time.Millisecond * 100)
				continue
			}
			log.Errorf("accept error: %v", err)
			return
		}
		log.Infof("backend %v connected", conn.RemoteAddr())
		go g.handleBackendConn(conn)
	}
}

func (g *Service) handleBackendConn(conn net.Conn) {
	endpoint, err := g.handShakeBackend(conn)
	if err != nil {
		log.Errorf("backend handshake: %v", err)
		conn.Close()
		return
	}
	defer log.Infof("backend %v(%v) disconnected", endpoint.NodeID(), endpoint.RemoteAddr())
	defer g.closeBackend(endpoint)
	endpoint.SetContext(g.Context())
	endpoint.Go(true, false)
	g.sendBackendToInstance(endpoint)

	var reader = bufio.NewReader(conn)
	var interval = time.Duration(g.Context().Env().NetPeerReadTimeout) * time.Second
	var encoder = codec.ServerProtocolCodec
	for !g.IsClosing() {
		var pkt = fatchoy.MakePacket()
		conn.SetReadDeadline(fatchoy.Now().Add(interval))
		if _, err := encoder.Unmarshal(reader, pkt); err != nil {
			if err != io.EOF {
				log.Errorf("backend %v read packet %v: %v", endpoint.NodeID(), pkt, err)
			}
			break
		}
		pkt.Endpoint = endpoint
		g.dispatchPacket(pkt)
	}
}

// backend握手: 服务注册
func (g *Service) handShakeBackend(conn net.Conn) (fatchoy.Endpoint, error) {
	var encoder = codec.ServerProtocolCodec
	var req protocol.RegisterReq
	var reqPkt = fatchoy.MakePacket()
	if err := qnet.ReadProtoMessage(conn, encoder, nil, reqPkt, &req); err != nil {
		log.Errorf("read register message: %v", err)
		return nil, err
	}

	rspPkt := fatchoy.MakePacket()
	rspPkt.Node = g.NodeID()
	rspPkt.Command = uint32(protocol.MSG_SM_INTERNAL_REGISTER)
	rspPkt.Seq = reqPkt.Seq

	// 验证安全性 AccessKeyID
	var env = g.Environ()
	var token = cluster.SignAccessToken(fatchoy.NodeID(req.Node), env.GameID, env.AccessKey)
	if req.AccessToken != token {
		log.Errorf("register token mismatch [%s] != [%s]", req.AccessToken, token)
		rspPkt.SetErrno(uint32(protocol.ErrRegistrationDenied))
		qnet.SendPacketMessage(conn, encoder, nil, rspPkt)
		return nil, eBackendRegisterDenied
	}

	// 是否已经注册
	var node = fatchoy.NodeID(req.Node)
	exist := g.backends.Get(node)
	if exist != nil {
		log.Errorf("duplicate registration of node %v", node)
		rspPkt.SetErrno(uint32(protocol.ErrDuplicateRegistration))
		qnet.SendPacketMessage(conn, encoder, nil, rspPkt)
		return nil, eBackendRegisterDenied
	}

	var ctx = g.Context()
	var endpoint = qnet.NewTcpConn(node, conn, encoder, nil, nil,
		ctx.Env().EndpointOutboundQueueSize, g.sStats)
	g.backends.Add(node, endpoint)

	var resp = &protocol.RegisterAck{
		Node: uint32(g.NodeID()),
	}
	rspPkt.Body = resp
	qnet.SendPacketMessage(conn, encoder, nil, rspPkt)

	log.Infof("backend %v registered", node)
	var notify = &protocol.InstanceStateNtf{
		State: protocol.StateUp,
		Peers: []uint32{uint32(node)},
	}
	g.broadcastToBackends(int32(protocol.MSG_INTERNAL_INSTANCE_STATE_NOTIFY), notify, node)
	return endpoint, nil
}

func (g *Service) NodeInfo() *protocol.NodeInfo {
	return &protocol.NodeInfo{
		Node:      uint32(g.NodeID()),
		Interface: g.saddr,
	}
}

func (g *Service) AddDependency(info *protocol.NodeInfo) {
	// do nothing
}

func (g *Service) DelDependency(removeAll bool, node fatchoy.NodeID) {
	// do nothing
}
