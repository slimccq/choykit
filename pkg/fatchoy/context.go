// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"sync"
	"sync/atomic"

	"devpkg.work/choykit/pkg/log"
	"github.com/pkg/errors"
)

// 服务的上下文
type ServiceContext struct {
	done       chan struct{}  //
	wg         sync.WaitGroup //
	closing    int32          //
	guard      sync.Mutex     // guard finalizers
	finalizers []func()       // finalizers
	service    Service        // service对象
	env        *Environ       // 环境变量
	opt        *Options       // 命令行参数
	router     *Router        // 路由对象
	inbound    chan *Packet   // 收取消息队列
	outbound   chan *Packet   // 发送消息队列
	filter     PacketFilter   // 消息过滤器(rpc使用)
}

func NewServiceContext(opt *Options, env *Environ) *ServiceContext {
	return &ServiceContext{
		opt:      opt,
		env:      env,
		done:     make(chan struct{}),
		inbound:  make(chan *Packet, env.ContextInboundQueueSize),
		outbound: make(chan *Packet, env.ContextOutboundQueueSize),
	}
}

func (c *ServiceContext) Start(srv Service) error {
	c.service = srv
	c.router = NewRouter(srv.NodeID())
	c.Go()
	log.Infof("start initialize %s service", srv.Name())
	if err := srv.Init(c); err != nil {
		return err
	}
	log.Infof("start run %s service %v", c.service.Name(), c.service.NodeID())
	if err := c.service.Startup(); err != nil {
		return err
	}
	return nil
}

func (c *ServiceContext) Options() *Options {
	return c.opt
}

func (c *ServiceContext) Env() *Environ {
	return c.env
}

func (c *ServiceContext) Service() Service {
	return c.service
}

func (c *ServiceContext) Router() *Router {
	return c.router
}

func (c *ServiceContext) finally() {
	defer Catch()
	c.guard.Lock()
	defer c.guard.Unlock()
	if cnt := len(c.finalizers); cnt > 0 {
		for _, f := range c.finalizers {
			f()
		}
		log.Infof("%d finalizers executed", cnt)
	}
}

func (c *ServiceContext) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) > 0
}

func (c *ServiceContext) Shutdown() {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		return
	}
	c.finally()
	log.Infof("start shutdown %s service", c.service.Name())
	c.service.Shutdown()
	log.Infof("%s service shutdown succeed", c.service.Name())
	close(c.done)
	c.wg.Wait()
	close(c.inbound)
	close(c.outbound)
	c.inbound = nil
	c.outbound = nil
	log.Infof("service context shutdown succeed")
}

func (c *ServiceContext) InboundQueue() chan<- *Packet {
	return c.inbound
}

func (c *ServiceContext) AddFinalizer(finalizer func()) {
	c.guard.Lock()
	c.finalizers = append(c.finalizers, finalizer)
	c.guard.Unlock()
}

func (c *ServiceContext) SendMessage(pkt *Packet) error {
	select {
	case c.outbound <- pkt:
		if n := len(c.outbound); n*3 >= cap(c.outbound)*2 {
			log.Warnf("ServiceContext: outbound message queue #%d is 2/3 full!", n)
		}
		return nil
	default:
		return errors.WithStack(ErrOutboundQueueOverflow)
	}
}

func (c *ServiceContext) SetMessageFilter(f PacketFilter) PacketFilter {
	old := c.filter
	c.filter = f
	return old
}

// filter和dispatch执行在不同的goroutine
func (c *ServiceContext) filterMessage(pkt *Packet) bool {
	defer Catch()
	if c.filter != nil {
		return c.filter(pkt)
	}
	return false
}

func (c *ServiceContext) Go() {
	c.wg.Add(2)
	go c.serve(0, c.inbound)
	go c.serve(1, c.outbound)
}

func (c *ServiceContext) serve(i int, queue <-chan *Packet) {
	defer c.wg.Done()
	defer log.Debugf("message dispatcher #%d stopped", i)
	log.Debugf("message dispatcher #%d start serving", i)
	for !c.IsClosing() {
		select {
		case pkt, ok := <-queue:
			if !ok {
				return
			}
			pkt.Endpoint.SetContext(c)
			if c.router.IsLoopBack(pkt) {
				if !c.filterMessage(pkt) {
					if err := c.service.Execute(pkt); err != nil {
						log.Errorf("dispatcher #%d execute packet [%v]: %v", i, pkt, err)
					}
				}
			} else {
				if err := c.router.Route(pkt); err != nil {
					log.Errorf("dispatcher #%d route packet [%v]: %v", i, pkt, err)
				}
			}

		case <-c.done:
			return
		}
	}
}
