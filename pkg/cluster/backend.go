// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"net"
	"sync"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"devpkg.work/choykit/pkg/qnet"
)

// 服务节点
type Backend struct {
	Node
	qnet.RpcFactory
	done       chan struct{}        //
	wg         sync.WaitGroup       //
	errors     chan error           // 网络错误
	stats      *fatchoy.Stats       // 收发数据包统计
	discovery  *EtcdDiscovery       // 服务发现
	listener   net.Listener         // 侦听器
	endpoints  *fatchoy.EndpointMap // 所有连接
	dependency []uint8              // 依赖的节点类型
	depNodes   NodeInfoMap          // 当前存在的依赖节点
}

func (s *Backend) Init(ctx *fatchoy.ServiceContext) error {
	if err := s.Node.Init(ctx); err != nil {
		return err
	}
	if err := s.RpcFactory.Init(ctx); err != nil {
		return err
	}
	s.done = make(chan struct{})
	s.errors = make(chan error, 16)
	s.stats = fatchoy.NewStats(qnet.NumStat)
	s.endpoints = fatchoy.NewEndpointMap()

	env := ctx.Env()
	dependency, err := DependencyServiceTypes(env.ServiceDependency)
	if err != nil {
		return err
	}
	s.dependency = dependency

	s.discovery = NewEtcdDiscovery(env, s)
	s.AddMessageHandler(true, s.handleMessage)
	return nil
}

func (s *Backend) Startup() error {
	if err := s.Node.Startup(); err != nil {
		return err
	}
	s.RpcFactory.Go()
	if err := s.startListen(); err != nil {
		return err
	}
	s.discovery.Start()

	s.wg.Add(1)
	go s.serveNetErr()
	return nil
}

func (s *Backend) Shutdown() {
	close(s.done)
	s.wg.Wait()
	for _, endpoint := range s.endpoints.List() {
		endpoint.ForceClose(nil)
	}
	s.depNodes.Clear()
	s.discovery.Close()
	s.Node.Shutdown()
	s.RpcFactory.Shutdown()
	s.errors = nil
	s.endpoints = nil
	s.stats = nil
}

func (s *Backend) IsMyDependency(node fatchoy.NodeID) bool {
	if s.node == node {
		return false
	}
	for _, srvType := range s.dependency {
		if srvType == node.Service() {
			return true
		}
	}
	return false
}

func (s *Backend) AddDependency(info *protocol.NodeInfo) {
	node := fatchoy.NodeID(info.Node)
	log.Debugf("dependency node alive: %s, %s", node, info.Interface)
	if !s.IsMyDependency(node) {
		return
	}
	s.depNodes.AddNode(info)
	endpoint := s.endpoints.Get(node)
	if endpoint == nil && info.Interface != "" {
		if err := s.establishTo(node, info.Interface); err != nil {
			log.Errorf("establish to node %v: %v", node, err)
		}
	}
}

func (s *Backend) DelDependency(etcdDown bool, node fatchoy.NodeID) {
	log.Debugf("dependency node lost: %v, %v", etcdDown, node)
	if etcdDown {
		s.depNodes.Clear()
		return
	}
	if s.IsMyDependency(node) {
		s.depNodes.DeleteNode(node)
	}
}

func (s *Backend) NodeInfo() *protocol.NodeInfo {
	return &protocol.NodeInfo{
		Node:      uint32(s.node),
		//Interface: s.Context().Options().Interface,
	}
}
