// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"sync"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"github.com/gogo/protobuf/proto"
)

type RpcHandler func(*RpcContext) error

// RPC上下文
type RpcContext struct {
	command  int32            // RPC请求消息ID
	reply    int32            // RPC相应消息ID
	errno    uint32           // 错误码
	body     interface{}      // 相应消息
	deadline time.Time        // 超时
	handler  RpcHandler       // 消息处理函数
	done     chan *RpcContext // Strobes when RPC is completed
}

func NewRpcContext(request, reply int32, body interface{}, handler RpcHandler) *RpcContext {
	return &RpcContext{
		command: request,
		reply:   reply,
		body:    body,
		handler: handler,
	}
}

func (r *RpcContext) Body() interface{} {
	return r.body
}

func (r *RpcContext) Errno() uint32 {
	return r.errno
}

func (r *RpcContext) Succeed() bool {
	return r.errno == 0
}

func (r *RpcContext) Run() error {
	if r.handler != nil {
		return r.handler(r)
	}
	return nil
}

func (r *RpcContext) DecodeMsg(v proto.Message) error {
	if r.body == nil {
		return nil
	}
	return fatchoy.DecodeAsMsg(r.body, v)
}

func (r *RpcContext) Done(ec uint32, body interface{}) {
	r.errno = ec
	r.body = body
	r.notify()
}

// 同步RPC在这里等待
func (r *RpcContext) notify() {
	if r.done != nil {
		select {
		case r.done <- r:
			// ok
		default:
			// We don't want to block here. It is the caller's responsibility to make
			// sure the channel has enough buffer space.
		}
	}
}

// RPC工厂
type RpcFactory struct {
	sync.Mutex
	done     chan struct{}           //
	wg       sync.WaitGroup          //
	pending  map[uint32]*RpcContext  // 待响应的RPC
	registry map[int32]bool          // 注册的响应消息
	seq      uint32                  // 序列号生成
	ttl      time.Duration           // 默认超时
	handler  fatchoy.PacketHandler   // RPC回调
	ctx      *fatchoy.ServiceContext // Context对象
}

func (r *RpcFactory) Init(ctx *fatchoy.ServiceContext) error {
	r.done = make(chan struct{})
	r.pending = make(map[uint32]*RpcContext)
	r.registry = make(map[int32]bool)
	r.ttl = time.Duration(ctx.Env().GetInt(fatchoy.NET_RPC_TTL)) * time.Second
	r.ctx = ctx
	r.seq = 2000 // magic number
	ctx.SetMessageFilter(r.filterRpcMessage)
	return nil
}

func (r *RpcFactory) Go() {
	r.wg.Add(1)
	go r.serve()
}

func (r *RpcFactory) Shutdown() {
	log.Debugf("start shutdown rpc factory")
	close(r.done)
	r.wg.Wait()
}

// 异步RPC
func (r *RpcFactory) CallAsync(request, reply int32, body proto.Message, cb RpcHandler) *RpcContext {
	if request == reply {
		log.Panicf("request[%d] should not equal to reply", request)
	}
	r.Lock()
	defer r.Unlock()
	var rpc = NewRpcContext(request, reply, body, cb)
	r.makeCall(rpc)
	return rpc
}

// 同步RPC
func (r *RpcFactory) Call(request, reply int32, body proto.Message) *RpcContext {
	if request == reply {
		log.Panicf("request[%d] should not equal to reply", request)
	}
	r.Lock()
	defer r.Unlock()
	var rpc = NewRpcContext(request, reply, body, nil)
	rpc.done = make(chan *RpcContext, 1)
	rpc = <-r.makeCall(rpc).done
	return rpc
}

func (r *RpcFactory) makeCall(ctx *RpcContext) *RpcContext {
	r.registry[ctx.reply] = true
	var seq = r.counter()
	ctx.deadline = fatchoy.Now().Add(r.ttl)
	var pkt = fatchoy.NewPacket(uint32(ctx.command), seq, fatchoy.PacketFlagRpc, ctx.body)
	r.ctx.SendMessage(pkt)
	r.pending[seq] = ctx
	return ctx
}

// 生成RPC序列号
func (r *RpcFactory) counter() uint32 {
	var seq = r.seq
	r.seq++ // we assume this id won't exhaust
	return seq
}

// 从接收消息中过滤RPC响应
func (r *RpcFactory) filterRpcMessage(pkt *fatchoy.Packet) bool {
	r.Lock()
	var command = int32(pkt.Command)
	var seq = pkt.Seq
	if !r.registry[command] {
		r.Unlock()
		return false
	}
	r.registry[command] = false
	rpc, found := r.pending[seq]
	if !found {
		r.Unlock()
		log.Errorf("unexpected RPC reply: %v", pkt)
		return true
	}
	delete(r.pending, seq)
	rpc.errno = pkt.Errno()
	rpc.body = pkt.Body
	r.Unlock()
	if err := r.ctx.Service().Execute(rpc); err != nil {
		log.Errorf("execute rpc #%d message %d: %v", seq, rpc.reply, err)
	}
	return true
}

func (r *RpcFactory) serve() {
	defer r.wg.Done()
	defer log.Debugf("rpc factory stop serving")
	log.Debugf("rpc factory start serving")
	var ticker = time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var expired = make([]*RpcContext, 0, 16)
	for {
		select {
		case now := <-ticker.C:
			r.Lock()
			for seq, rpc := range r.pending {
				if now.After(rpc.deadline) {
					delete(r.pending, seq)
					expired = append(expired, rpc)
				}
			}
			r.Unlock()
			for _, rpc := range expired {
				rpc.Done(uint32(protocol.ErrRpcTimeout), nil)
				if err := r.ctx.Service().Execute(rpc); err != nil {
					log.Errorf("execute rpc reply %d: %v", rpc.reply, err)
				}
			}
			expired = expired[:0]

		case <-r.done:
			return
		}
	}
}
