// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"errors"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"github.com/golang/protobuf/proto"
)

var (
	eBackendRegisterDenied = errors.New("backend registration denied")
	errInvalidInterface    = errors.New("invalid interface address")
)

func (g *Service) dispatchPacket(pkt *fatchoy.Packet) {
	if g.router.IsLoopBack(pkt) {
		if err := g.handleMessage(pkt); err != nil {
			log.Errorf("dispatch message: %v, %v", pkt, err)
		}
	} else {
		if err := g.router.Route(pkt); err != nil {
			log.Errorf("route message %v, %v", pkt, err)
		}
	}
}

func (g *Service) propagateClientLost(session fatchoy.Endpoint) {
	ud := GetSessionUData(session)
	ns := g.router.GetSubNodes(int32(protocol.MSG_CLIENT_LOST_NOTIFY), int32(protocol.MSG_CLIENT_LOST_NOTIFY))
	for _, node := range ns {
		var notify = &protocol.ClientLostNtf{
			UserId:  ud.userid,
			Session: uint32(session.NodeID()),
		}
		g.backends.Get(node)
		g.broadcastToBackends(int32(protocol.MSG_CLIENT_LOST_NOTIFY), notify, 0)
	}
}

// 关闭客户端session
func (g *Service) closeSession(session fatchoy.Endpoint) {
	g.sessions.Delete(session.NodeID())
	session.ForceClose(nil)
	g.propagateClientLost(session)
}

// 关闭后端服务连接
func (g *Service) closeBackend(endpoint fatchoy.Endpoint) {
	var node = endpoint.NodeID()
	g.backends.Delete(node)
	var notify = &protocol.InstanceStateNtf{
		State: protocol.StateDown,
		Peers: []uint32{uint32(node)},
	}
	g.broadcastToBackends(int32(protocol.MSG_INTERNAL_INSTANCE_STATE_NOTIFY), notify, node)
	endpoint.ForceClose(nil)
}

// 广播给所有backend
func (g *Service) broadcastToBackends(command int32, notify proto.Message, except fatchoy.NodeID) {
	for _, endpoint := range g.backends.List() {
		if endpoint.NodeID() != except {
			pkt := fatchoy.NewPacket(g.NodeID(), uint32(command), 0, 0, notify)
			endpoint.SendPacket(pkt)
		}
	}
}

// 发送当前所有服务给新加入节点
func (g *Service) sendBackendToInstance(endpoint fatchoy.Endpoint) {
	var notify = &protocol.InstanceStateNtf{
		State: protocol.StateUp,
	}
	for _, ses := range g.backends.List() {
		if ses.NodeID() != endpoint.NodeID() {
			notify.Peers = append(notify.Peers, uint32(ses.NodeID()))
		}
	}
	if len(notify.Peers) > 0 {
		pkt := fatchoy.NewPacket(g.NodeID(), uint32(protocol.MSG_INTERNAL_INSTANCE_STATE_NOTIFY), 0, 0, notify)
		endpoint.SendPacket(pkt)
	}
}

// 主动关闭
func (g *Service) kick(session fatchoy.Endpoint, reason uint32, propagate bool) {
	var notify = &protocol.ClientDisconnectNtf{
		Reason: reason,
	}
	pkt := fatchoy.NewPacket(g.NodeID(), uint32(protocol.MSG_CLIENT_DISCONNECT_NOTIFY), 0, 0, notify)
	session.SendPacket(pkt)
	session.Close()
	if propagate {
		g.propagateClientLost(session)
	}
}

func (g *Service) disconnectAll() {
	for _, session := range g.sessions.List() {
		g.kick(session, uint32(protocol.ErrServerMaintenance), false)
	}
}
