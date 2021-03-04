// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/protocol"
)

type RoutePolicy struct {
	backends *fatchoy.EndpointMap
	sessions *fatchoy.EndpointMap
}

// client/backend --> gate:   node == gate
func (p *RoutePolicy) IsLoopBack(router *fatchoy.Router, pkt *fatchoy.Packet) bool {
	if pkt.Node == 0 {
		if !pkt.Node.IsBackend() {
			hasSub := router.HasSubNodes(int32(protocol.MSG_CLIENT_START_ID), int32(protocol.MSG_CLIENT_END_ID))
			if hasSub {
				return false
			}
		}
		return true
	}
	return pkt.Node == router.NodeID()
}

// client --> backend:	node == backend and refer == session
// backend --> backend: node == backend
// backend --> client:  node == session
func (p *RoutePolicy) Lookup(router *fatchoy.Router, pkt *fatchoy.Packet) fatchoy.Endpoint {
	var from = pkt.Endpoint.NodeID()
	if pkt.Node.IsBackend() {
		var dest = router.GetEntry(pkt.Node)
		if endpoint := p.backends.Get(dest); endpoint != nil {
			pkt.Node = from
			pkt.Endpoint = endpoint
			return endpoint
		}
	} else {
		if session := p.sessions.Get(pkt.Node); session != nil {
			pkt.Node = from
			pkt.Endpoint = session
			return session
		}
	}
	return nil
}

// 广播
func (p *RoutePolicy) Multicast(router *fatchoy.Router, pkt *fatchoy.Packet) bool {
	switch {
	case pkt.Node == router.NodeID(): // 广播给所有客户端
		var from = pkt.Endpoint.NodeID()
		for _, session := range p.sessions.List() {
			var copy = pkt.Clone()
			copy.Node = from
			copy.Endpoint = session
			session.SendPacket(copy)
		}
		return true
	}
	return false
}

func (g *Service) initRouter() {
	g.router = fatchoy.NewRouter(g.NodeID())
	policy := &RoutePolicy{g.backends, g.sessions}
	g.router.AddPolicy(policy)
	g.router.AddPolicy(fatchoy.NewBasicRoutePolicy(g.backends))
}
