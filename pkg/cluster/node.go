// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"sync/atomic"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/log"
)

// 节点
type Node struct {
	choykit.Executor
	closing  int32                   //
	node     choykit.NodeID          // 节点编号
	codec    choykit.Codec           // 消息编解码
	handlers []choykit.PacketHandler // 消息处理函数
	ctx      *choykit.ServiceContext // context对象
}

func (s *Node) Init(ctx *choykit.ServiceContext) error {
	var env = ctx.Env()
	if err := s.Executor.Init(env.ExecutorCapacity, env.ExecutorConcurrency); err != nil {
		return err
	}
	s.codec = codec.NewServerCodec()
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
	s.codec = nil
	s.ctx = nil
	s.handlers = nil
}

func (s *Node) NodeID() choykit.NodeID {
	return s.node
}

func (s *Node) SetNodeID(v choykit.NodeID) {
	s.node = v
}

func (s *Node) IsClosing() bool {
	return atomic.LoadInt32(&s.closing) > 0
}

func (s *Node) Codec() choykit.Codec {
	return s.codec.Clone()
}

func (s *Node) Environ() *choykit.Environ {
	return s.ctx.Env()
}

func (s *Node) Context() *choykit.ServiceContext {
	return s.ctx
}

func (s *Node) SendPacket(pkt *choykit.Packet) error {
	return s.ctx.SendMessage(pkt)
}

// 添加消息处理函数
func (s *Node) AddMessageHandler(isPrepend bool, f choykit.PacketHandler) {
	if isPrepend {
		s.handlers = append([]choykit.PacketHandler{f}, s.handlers...)
	} else {
		s.handlers = append(s.handlers, f)
	}
}

// 执行消息处理
func (s *Node) Dispatch(pkt *choykit.Packet) error {
	var err error
	for _, f := range s.handlers {
		if er := f(pkt); er != nil {
			err = er
			log.Errorf("dispatch message (%v): %v", pkt, er)
		}
	}
	return err
}
