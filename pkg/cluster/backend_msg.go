// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
)

func (s *Backend) handleMessage(pkt *choykit.Packet) error {
	switch protocol.InternalMsgType(pkt.Command) {
	case protocol.MSG_INTERNAL_KEEP_ALIVE_STATUS:
		return s.handlePong(pkt)

	case protocol.MSG_INTERNAL_INSTANCE_STATE_NOTIFY:
		return s.handleInstanceStateNtf(pkt)
	}
	return nil
}

func (s *Backend) handlePong(pkt *choykit.Packet) error {
	var msg protocol.KeepAliveAck
	if err := pkt.DecodeMsg(&msg); err != nil {
		return err
	}
	log.Debugf("recv pong %d from %v", msg.Time, pkt.Endpoint.NodeID())
	return nil
}

func (s *Backend) handleInstanceStateNtf(pkt *choykit.Packet) error {
	var msg protocol.InstanceStateNtf
	if err := pkt.DecodeMsg(&msg); err != nil {
		return err
	}
	// TODO:
	return nil
}

// 其它节点注册进来
func (s *Backend) handleRegister(req *protocol.RegisterReq, pkt *choykit.Packet) bool {
	var env = s.Environ()
	var token = SignAccessToken(choykit.NodeID(req.Node), env.GameID, env.AccessKey)
	if req.AccessToken != token {
		log.Errorf("register token mismatch [%s] != [%s]", req.AccessToken, token)
		pkt.SetErrno(uint32(protocol.ErrRegistrationDenied))
		return false
	}
	// 是否已经注册
	var node = choykit.NodeID(req.Node)
	exist := s.endpoints.Get(node)
	if exist != nil {
		log.Errorf("duplicate registration of node %v", node)
		pkt.SetErrno(uint32(protocol.ErrDuplicateRegistration))
		return false
	}
	return true
}
