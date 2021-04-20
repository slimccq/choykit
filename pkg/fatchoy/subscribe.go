// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "sync"

// 订阅了消息的节点
type MessageSubscription struct {
	StartID int32
	EndID   int32
	Nodes   NodeIDSet
}

func (s *MessageSubscription) Match(id int32) bool {
	return s.StartID == id && id == s.EndID
}

// 假定条件`startId <= endID`已经成立
func (s *MessageSubscription) MatchRange(startId, endId int32) bool {
	return startId <= s.StartID && s.EndID <= endId
}

// 服务节点消息订阅
// 设计为支持跨服务转发消息的使用场景，例如game向gate订阅了登录消息，gate才会把登录转发给game
type MessageSubscriber struct {
	sync.RWMutex
	subList []*MessageSubscription
}

func NewMessageSub() *MessageSubscriber {
	return &MessageSubscriber{
		subList: make([]*MessageSubscription, 0, 8),
	}
}

// 获取订阅了某个消息的节点
func (s *MessageSubscriber) GetSubscribeNodesOf(msgID int32) NodeIDSet {
	s.RLock()
	var nodes NodeIDSet
	for _, sub := range s.subList {
		if sub.Match(msgID) {
			nodes = append(nodes, sub.Nodes...)
		}
	}
	s.RUnlock()
	return nodes
}

// 获取订阅了某个消息区间的节点
func (s *MessageSubscriber) GetSubscribeNodes(startId, endId int32) NodeIDSet {
	s.RLock()
	var nodes NodeIDSet
	for _, sub := range s.subList {
		if sub.MatchRange(startId, endId) {
			nodes = append(nodes, sub.Nodes...)
		}
	}
	s.RUnlock()
	return nodes
}

// 添加订阅
func (s *MessageSubscriber) AddSubNode(msgId int32, node NodeID) {
	s.Lock()
	for _, sub := range s.subList {
		if sub.Match(msgId) {
			if sub.Nodes.Has(node) >= 0 {
				s.Unlock()
				return
			}
		}
	}
	var sub = &MessageSubscription{
		StartID: msgId,
		EndID:   msgId,
	}
	sub.Nodes = sub.Nodes.Insert(node)
	s.subList = append(s.subList, sub)
	s.Unlock()
}

// 添加订阅（区间）
func (s *MessageSubscriber) AddSubNodeRange(startId, endId int32, node NodeID) {
	s.Lock()
	for _, sub := range s.subList {
		if sub.MatchRange(startId, endId) {
			if sub.Nodes.Has(node) >= 0 {
				s.Unlock()
				return
			}
		}
	}
	var sub = &MessageSubscription{
		StartID: startId,
		EndID:   endId,
	}
	sub.Nodes = sub.Nodes.Insert(node)
	s.subList = append(s.subList, sub)
	s.Unlock()
}

// 删除一个节点的订阅信息
func (s *MessageSubscriber) DeleteNodeSubscription(node NodeID) {
	s.Lock()
	for i, sub := range s.subList {
		sub.Nodes = sub.Nodes.Delete(node)
		if len(sub.Nodes) == 0 {
			s.subList = append(s.subList[:i], s.subList[i+1:]...)
		}
	}
	s.Unlock()
}
