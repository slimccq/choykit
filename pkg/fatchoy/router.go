// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"sync"
)

// 路由策略
type RoutePolicy interface {
	IsLoopBack(*Router, *Packet) bool
	Multicast(*Router, *Packet) bool
	Lookup(*Router, *Packet) Endpoint
}

// 路由表
type RoutingTableEntry struct {
	src, dest NodeID
}

type RoutingTable struct {
	sync.RWMutex
	entries map[NodeID]NodeID
}

func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		entries: make(map[NodeID]NodeID),
	}
}

func (r *RoutingTable) GetEntry(key NodeID) NodeID {
	r.RLock()
	v := r.entries[key]
	r.RUnlock()
	return v
}

func (r *RoutingTable) AddEntry(src, dst NodeID) {
	r.Lock()
	r.entries[src] = dst
	r.Unlock()
}

func (r *RoutingTable) DeleteEntry(src NodeID) {
	r.Lock()
	delete(r.entries, src)
	r.Unlock()
}

func (r *RoutingTable) DeleteDestEntry(dest NodeID) {
	r.Lock()
	for src, dst := range r.entries {
		if dest == dst {
			delete(r.entries, src)
		}
	}
	r.Unlock()
}

func (r *RoutingTable) EntryList() []RoutingTableEntry {
	r.RLock()
	var entries = make([]RoutingTableEntry, 0, len(r.entries))
	for k, v := range r.entries {
		entries = append(entries, RoutingTableEntry{k, v})
	}
	r.RUnlock()
	return entries
}

// 消息订阅
type MessageSubscriber struct {
	sync.RWMutex
	byRange map[int64]NodeIDSet
}

func NewMessageSub() *MessageSubscriber {
	return &MessageSubscriber{
		byRange: make(map[int64]NodeIDSet),
	}
}

func (s *MessageSubscriber) GetSubNodes(startMsg, endMsg int32) NodeIDSet {
	s.RLock()
	key := (int64(startMsg) << 32) | int64(endMsg)
	nodes := s.byRange[key]
	s.RUnlock()
	return nodes
}

func (s *MessageSubscriber) HasSubNodes(startMsg, endMsg int32) bool {
	s.RLock()
	key := (int64(startMsg) << 32) | int64(endMsg)
	_, found := s.byRange[key]
	s.RUnlock()
	return found
}

func (s *MessageSubscriber) AddSubNode(start, end int32, node NodeID) {
	s.Lock()
	key := (int64(start) << 32) | int64(end)
	s.byRange[key] = s.byRange[key].Add(node)
	s.Unlock()
}

func (s *MessageSubscriber) DeleteSubNode(start, end int32, node NodeID) {
	s.Lock()
	key := (int64(start) << 32) | int64(end)
	s.byRange[key] = s.byRange[key].Delete(node)
	s.Unlock()
}

func (s *MessageSubscriber) DeleteNodeSubs(dest NodeID) {
	s.Lock()
	for k, ns := range s.byRange {
		if ns.Has(dest) >= 0 {
			s.byRange[k] = s.byRange[k].Delete(dest)
		}
	}
	s.Unlock()
}

// 路由器
type Router struct {
	*MessageSubscriber               // 消息订阅
	*RoutingTable                    // 路由表
	node               NodeID        // 节点号
	policies           []RoutePolicy // 路由策略
}

func NewRouter(node NodeID) *Router {
	return &Router{
		node:              node,
		MessageSubscriber: NewMessageSub(),
		RoutingTable:      NewRoutingTable(),
	}
}

func (r *Router) AddPolicy(policy RoutePolicy) {
	r.policies = append(r.policies, policy)
}

func (r *Router) NodeID() NodeID {
	return r.node
}

func (r *Router) IsLoopBack(pkt *Packet) bool {
	if pkt.Endpoint == nil { // endpoint为nil则是本地消息
		return true
	}
	for _, policy := range r.policies {
		if policy.IsLoopBack(r, pkt) {
			return true
		}
	}
	return true // last choice
}

func (r *Router) Route(pkt *Packet) error {
	for _, policy := range r.policies {
		if policy.Multicast(r, pkt) {
			break
		}
		if endpoint := policy.Lookup(r, pkt); endpoint != nil {
			return endpoint.SendPacket(pkt)
		}
	}
	return ErrDestinationNotReachable
}
