// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"devpkg.work/choykit/pkg/qnet"
)

func (g *Service) handleMessage(pkt *fatchoy.Packet) error {
	defer fatchoy.Catch()
	switch int32(pkt.Command) {
	case int32(protocol.MSG_CM_INTERNAL_KEEP_ALIVE):
		return g.handleKeepAlive(pkt)
	case int32(protocol.MSG_CM_CLIENT_PING):
		return g.handlePingReq(pkt)
	case int32(protocol.MSG_CM_KICK_CLIENT):
		return g.handleKickSession(pkt)
	case int32(protocol.MSG_CM_INTERNAL_SUBSCRIBE):
		return g.handleSubscribe(pkt)
	case int32(protocol.MSG_SM_FORWARD_CLIENT_PACKET):
		return g.handleForwardMsgBack(pkt)
	}
	return nil
}

// backend心跳
func (g *Service) handleKeepAlive(pkt *fatchoy.Packet) error {
	var req protocol.KeepAliveReq
	if err := pkt.DecodeMsg(&req); err != nil {
		return err
	}
	log.Debugf("recv ping %d from %v", req.Time, pkt.Endpoint.NodeID())
	resp := &protocol.KeepAliveAck{
		Time: fatchoy.Now().Unix(),
	}
	return pkt.Reply(resp)
}

// backend订阅它想处理的消息
func (g *Service) handleSubscribe(pkt *fatchoy.Packet) error {
	var req protocol.SubscribeReq
	if err := pkt.DecodeMsg(&req); err != nil {
		return err
	}
	var node = pkt.Endpoint.NodeID()
	if req.MsgEnd < req.MsgStart {
		req.MsgEnd, req.MsgStart = req.MsgStart, req.MsgEnd
	}
	g.router.AddSubNode(req.MsgStart, req.MsgEnd, node)
	var resp protocol.SubscribeReq
	return pkt.Reply(&resp)
}

// 踢client下线
func (g *Service) handleKickSession(pkt *fatchoy.Packet) error {
	var req protocol.KickClientReq
	if err := pkt.DecodeMsg(&req); err != nil {
		return err
	}
	var resp protocol.KickClientAck
	for _, sid := range req.Sessions {
		if session := g.sessions.Get(fatchoy.NodeID(sid)); session != nil {
			g.kick(session, req.Reason, true)
			resp.Count++
		}
	}
	return pkt.Reply(&resp)
}

// client心跳
func (g *Service) handlePingReq(pkt *fatchoy.Packet) error {
	var req protocol.ClientPingReq
	if err := pkt.DecodeMsg(&req); err != nil {
		return err
	}
	resp := &protocol.ClientPongAck{
		Time: fatchoy.Now().Unix(),
	}
	log.Debugf("recv ping %d from %v", req.Time, pkt.Endpoint.NodeID())
	return pkt.Reply(resp)
}

// backend返回结果
func (g *Service) handleForwardMsgBack(pkt *fatchoy.Packet) error {
	var msg protocol.ForwardClientMsg
	if err := pkt.DecodeMsg(&msg); err != nil {
		return err
	}
	var session = g.sessions.Get(fatchoy.NodeID(msg.Session))
	if session == nil {
		log.Errorf("session %d not exist at internal login ack", session)
		return nil
	}
	cliPkt := fatchoy.NewPacket(g.NodeID(), msg.MsgId, 0, pkt.Seq, msg.MsgData)
	// 登录成功
	ud := GetSessionUData(session)
	ud.userid = msg.UserId
	ud.session = msg.Session
	g.sessions.Add(session.NodeID(), session)
	return session.SendPacket(cliPkt)
}

// 登录流程:
//  第1步: client ---> gate, ClientLoginReq
//  第2步: gate ---> game, LoginInternalReq
//  第3步: game ---> gate, LoginInternalAck
//  第4步: gate ---> client, ClientLoginAck
func (g *Service) forwardClientLogin(session fatchoy.Endpoint) error {
	var encoder = codec.ClientProtocolCodec
	var pkt = fatchoy.MakePacket()
	if err := qnet.ReadPacketMessage(session.RawConn(), encoder, nil, pkt); err != nil {
		log.Errorf("read login request: %v", err)
		return err
	}
	respPkt := fatchoy.NewPacket(session.NodeID(), uint32(protocol.MSG_SM_LOGIN), 0, pkt.Seq, nil)
	ns := g.router.GetSubNodes(int32(protocol.MSG_CLIENT_START_ID), int32(protocol.MSG_CLIENT_END_ID))
	if len(ns) == 0 {
		log.Errorf("no service can handle login")
		respPkt.SetErrno(uint32(protocol.ErrServiceNotAvailable))
		return session.SendPacket(respPkt)
	}
	endpoint := g.backends.Get(ns[0])
	if endpoint == nil {
		log.Errorf("backend %v not reachable", ns[0])
		respPkt.SetErrno(uint32(protocol.ErrServiceNotAvailable))
		return session.SendPacket(respPkt)
	}

	msgData, _ := pkt.EncodeBody()
	var fwdMsg = protocol.ForwardClientMsg{
		Session: uint32(session.NodeID()),
		MsgId:   pkt.Command,
		MsgData: msgData,
	}
	fwdPkt := fatchoy.NewPacket(ns[0], uint32(protocol.MSG_CM_FORWARD_CLIENT_PACKET), 0, pkt.Seq, &fwdMsg)
	return endpoint.SendPacket(fwdPkt)
}
