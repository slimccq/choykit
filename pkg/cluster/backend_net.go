// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"net"
	"strings"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"devpkg.work/choykit/pkg/qnet"
)

// 侦听其它node连接
func (s *Backend) startListen() error {
	opts := s.Context().Options()
	if opts.Interface == "" {
		return nil
	}
	var addr string
	addrs := strings.Split(opts.Interface, ",")
	switch len(addrs) {
	case 0:
		return nil // 没有监听地址，服务不需要互联
	case 1:
		addr = addrs[0]
	default:
		addr = addrs[len(addrs)-1] // 如果有多个地址，则选择最后一个地址
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = ln
	s.wg.Add(1)
	go s.serveAccept()
	return nil
}

func (s *Backend) serveAccept() {
	defer s.wg.Done()
	for !s.IsClosing() {
		conn, err := s.listener.Accept()
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
		go s.handleNodeAccept(conn)
	}
}

// 处理节点注册
func (s *Backend) handleNodeAccept(conn net.Conn) {
	var req protocol.RegisterReq
	var pkt1 = fatchoy.MakePacket()
	if err := qnet.ReadProtoMessage(conn, s.encoder, nil, pkt1, &req); err != nil {
		log.Errorf("read registration message: %v", err)
		return
	}
	pkt2 := fatchoy.MakePacket()
	pkt2.Node = s.NodeID()
	pkt2.Command = uint32(protocol.MSG_SM_INTERNAL_REGISTER)
	pkt2.Seq = pkt1.Seq
	pkt2.Body = &protocol.RegisterAck{Node: uint32(s.NodeID())}
	regOK := s.handleRegister(&req, pkt2)
	if err := qnet.SendPacketMessage(conn, s.encoder, nil, pkt2); err != nil {
		log.Errorf("send registration message: %v", err)
		return
	}
	if !regOK {
		return
	}
	var ctx = s.Context()
	var node = fatchoy.NodeID(req.Node)
	var endpoint = qnet.NewTcpConn(node, conn, s.encoder, nil, nil,
		ctx.Env().EndpointOutboundQueueSize, s.stats)
	s.endpoints.Add(node, endpoint)
	endpoint.Go(true, true)
	s.endpoints.Add(node, endpoint)
	s.Context().Router().AddEntry(node, node)
}

func (s *Backend) serveNetErr() {
	defer s.wg.Done()
	for !s.IsClosing() {
		select {
		case er := <-s.errors:
			e, ok := er.(*qnet.Error)
			if !ok {
				log.Errorf("unrecognized err %T: %v", er, er)
				return
			}
			var node = e.Endpoint.NodeID()
			e.Endpoint.ForceClose(nil)
			s.endpoints.Delete(node)
			s.depNodes.DeleteNode(node)
			s.Context().Router().DeleteEntry(node)

		case <-s.done:
			return
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// 连接到其它node
func (s *Backend) establishTo(node fatchoy.NodeID, addr string) error {
	if endpoint := s.endpoints.Get(node); endpoint != nil {
		log.Warnf("node %v already established", node)
		return nil
	}
	log.Infof("start connect node %v(%s)", node, addr)
	conn, err := net.DialTimeout("tcp", addr, 7*time.Second)
	if err != nil {
		return err
	}
	ctx := s.Context()
	endpoint := qnet.NewTcpConn(node, conn, s.encoder, s.errors, ctx.InboundQueue(),
		ctx.Env().EndpointOutboundQueueSize, s.stats)
	if err := s.register(endpoint); err != nil {
		return err
	}
	return nil
}

// 注册自己
func (s *Backend) register(endpoint fatchoy.Endpoint) error {
	var env = s.Environ()
	var token = SignAccessToken(s.node, env.GameID, env.AccessKey)
	var req = &protocol.RegisterReq{
		Node:        uint32(s.node),
		AccessToken: token,
	}
	log.Infof("start register self(%v) to node %v", s.node, endpoint.NodeID())
	var resp protocol.RegisterAck
	if err := qnet.RequestMessage(endpoint.RawConn(), s.encoder, nil, int32(protocol.MSG_CM_INTERNAL_REGISTER),
		req, &resp); err != nil {
		return err
	}

	var node = fatchoy.NodeID(resp.Node)
	endpoint.SetNodeID(node)
	endpoint.Go(true, true)
	s.endpoints.Add(node, endpoint)
	s.Context().Router().AddEntry(node, node)
	s.wg.Add(1)
	go s.servePing(endpoint)

	log.Infof("register to node %v succeed", endpoint.NodeID())
	return nil
}

// 发送心跳包
func (s *Backend) sendPing(now time.Time, endpoint fatchoy.Endpoint) {
	var msg = &protocol.KeepAliveReq{
		Time: now.Unix(),
	}
	pkt := fatchoy.NewPacket(endpoint.NodeID(), uint32(protocol.MSG_CM_INTERNAL_KEEP_ALIVE), 0, 0, msg)
	if err := endpoint.SendPacket(pkt); err != nil {
		log.Errorf("Send message %d: %v", pkt.Command, err)
	}
}

// 持续心跳
func (s *Backend) servePing(endpoint fatchoy.Endpoint) {
	defer s.wg.Done()
	defer log.Debugf("service endpoint %v stop serving", endpoint.NodeID())
	log.Debugf("start serve pinging for endpoint %v", endpoint.NodeID())

	var ctx = s.Context()
	ticker := time.NewTicker(time.Duration(ctx.Env().NetPeerPingInterval) * time.Second)
	defer ticker.Stop()

	s.sendPing(time.Now(), endpoint)
	for !s.IsClosing() {
		select {
		case now := <-ticker.C:
			if endpoint.IsClosing() {
				return
			}
			s.sendPing(now, endpoint)

		case <-s.done:
			return
		}
	}
}
