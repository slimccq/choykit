// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"net"
	"strings"
	"sync"

	"devpkg.work/choykit/pkg/cluster"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
)

type Service struct {
	cluster.Node
	done        chan struct{}          //
	wg          sync.WaitGroup         //
	closing     int32                  //
	discovery   *cluster.EtcdDiscovery //
	backendAddr fatchoy.NetInterface   // backend侦听地址
	sListener   net.Listener           // 与backend的连接
	cListener   net.Listener           // 与client的连接(TCP)
	wsvr        *WsServer              // 与client的连接(Websocket)
	backends    *fatchoy.EndpointMap   // 所有backend连接
	sessions    *fatchoy.EndpointMap   // 所有TCP client连接
	router      *fatchoy.Router        // 消息路由
	cStats      *fatchoy.Stats         // client消息统计
	sStats      *fatchoy.Stats         // backend消息统计
	nextSid     uint32                 // session id分配号
	pcu         uint32                 // PCU
}

func (g *Service) Init(ctx *fatchoy.ServiceContext) error {
	if err := g.Node.Init(ctx); err != nil {
		return err
	}
	g.nextSid = 1000
	g.done = make(chan struct{})
	g.cStats = fatchoy.NewStats(NumStat)
	g.sStats = fatchoy.NewStats(NumStat)
	g.backends = fatchoy.NewEndpointMap()
	g.sessions = fatchoy.NewEndpointMap()
	g.initRouter()

	env := ctx.Env()
	g.discovery = cluster.NewEtcdDiscovery(env, g)

	// 第一个地址监听server连接，其他地址监听client连接
	if len(env.NetInterfaces) < 2 {
		log.Errorf("invalid interfaces [%v] specified", env.NetInterfaces)
		return errInvalidInterface
	}
	g.backendAddr = fatchoy.NetInterface(*env.NetInterfaces[0])
	if err := g.createBackendListener(g.backendAddr); err != nil {
		return err
	}
	for i := 1; i < len(env.NetInterfaces); i++ {
		addr := fatchoy.NetInterface(*env.NetInterfaces[i])
		if strings.HasPrefix(addr.BindAddr, "ws") {
			if err := g.createWSServer(addr.BindAddr); err != nil {
				return err
			}
		} else {
			if err := g.createClientListener(addr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Service) Startup() error {
	if err := g.Node.Startup(); err != nil {
		return err
	}
	g.wg.Add(1)
	go g.serveBackend()

	if g.cListener != nil {
		g.wg.Add(1)
		go g.serveClientSession(g.cListener)
	}
	if g.wsvr != nil {
		g.wsvr.Start()
	}
	g.discovery.Start()
	return nil
}

func (g *Service) Shutdown() {
	g.disconnectAll()
	if g.cListener != nil {
		g.cListener.Close()
	}
	if g.wsvr != nil {
		g.wsvr.Close()
	}
	g.discovery.Close()
	g.sListener.Close()
	g.closing = 1
	close(g.done)
	g.wg.Wait()
	g.Node.Shutdown()
	g.sessions = nil
	g.backends = nil
	g.cListener = nil
	g.wsvr = nil
	g.router = nil
	g.cStats = nil
	g.sStats = nil
}

func (g *Service) IsClosing() bool {
	return g.closing > 0
}

func (g *Service) Name() string {
	return "gate"
}

func (g *Service) ID() uint8 {
	return protocol.SERVICE_GATEWAY
}
