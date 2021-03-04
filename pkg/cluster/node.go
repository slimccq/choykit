// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"sync/atomic"

	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
)

// 节点
type Node struct {
	fatchoy.Executor
	closing  int32                   //
	node     fatchoy.NodeID          // 节点编号
	encoder  fatchoy.ProtocolCodec   // 消息编解码
	handlers []fatchoy.PacketHandler // 消息处理函数
	ctx      *fatchoy.ServiceContext // context对象
}

func (s *Node) Init(ctx *fatchoy.ServiceContext) error {
	var env = ctx.Env()
	if err := s.Executor.Init(env.ExecutorCapacity, env.ExecutorConcurrency); err != nil {
		return err
	}
	s.encoder = codec.NewServerProtocolCodec()
	s.ctx = ctx
	return nil
}

func (s *Node) Startup() error {
	s.Executor.Go()
	return nil
}

func (s *Node) Shutdown() {
	s.Executor.Shutdown()
	log.Infof("executor shutdown succeed")
	s.encoder = nil
	s.ctx = nil
	s.handlers = nil
}

func (s *Node) NodeID() fatchoy.NodeID {
	return s.node
}

func (s *Node) SetNodeID(v fatchoy.NodeID) {
	s.node = v
}

func (s *Node) IsClosing() bool {
	return atomic.LoadInt32(&s.closing) > 0
}

func (s *Node) Environ() *fatchoy.Environ {
	return s.ctx.Env()
}

func (s *Node) Context() *fatchoy.ServiceContext {
	return s.ctx
}

func (s *Node) SendPacket(pkt *fatchoy.Packet) error {
	return s.ctx.SendMessage(pkt)
}

// 添加消息处理函数
func (s *Node) AddMessageHandler(isPrepend bool, f fatchoy.PacketHandler) {
	if isPrepend {
		s.handlers = append([]fatchoy.PacketHandler{f}, s.handlers...)
	} else {
		s.handlers = append(s.handlers, f)
	}
}

// 执行消息处理
func (s *Node) Dispatch(pkt *fatchoy.Packet) error {
	var err error
	for _, f := range s.handlers {
		if er := f(pkt); er != nil {
			err = er
			log.Errorf("dispatch message (%v): %v", pkt, er)
		}
	}
	return err
}
