// Copyright © 2021-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package cluster

import (
	"math"
	"sync"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/protocol"
)

// 按服务类型区分的节点信息
type NodeInfoMap struct {
	sync.RWMutex
	nodes [math.MaxUint8][]*protocol.NodeInfo
}

func NewNodeInfoMap() *NodeInfoMap {
	return &NodeInfoMap{}
}

// 所有本类型的节点
func (m *NodeInfoMap) GetNodes(srvType uint8) []*protocol.NodeInfo {
	m.RLock()
	v := m.nodes[srvType]
	m.RUnlock()
	return v
}

// 本区服所有类型的节点
func (m *NodeInfoMap) GetNodesBy(srvType uint8, district uint16) []*protocol.NodeInfo {
	m.RLock()
	var result []*protocol.NodeInfo
	for _, v := range m.nodes[srvType] {
		if choykit.NodeID(v.Node).District() == district {
			result = append(result, v)
		}
	}
	m.RUnlock()
	return result
}

// 添加一个节点
func (m *NodeInfoMap) AddNode(info *protocol.NodeInfo) {
	m.Lock()
	node := choykit.NodeID(info.Node)
	slice := m.nodes[node.Service()]
	for i, v := range slice {
		if v.Node == info.Node {
			slice[i] = info
			m.Unlock()
			return
		}
	}
	m.nodes[node.Service()] = append(m.nodes[node.Service()], info)
	m.Unlock()
}

func (m *NodeInfoMap) Clear() {
	m.Lock()
	for i := 0; i < len(m.nodes); i++ {
		m.nodes[i] = nil
	}
	m.Unlock()
}

// 删除某一类型的所有节点
func (m *NodeInfoMap) DeleteService(srvType uint8) {
	m.Lock()
	m.nodes[srvType] = nil
	m.Unlock()
}

// 删除一个节点
func (m *NodeInfoMap) DeleteNode(node choykit.NodeID) {
	m.Lock()
	slice := m.nodes[node.Service()]
	for i, v := range slice {
		if v.Node == uint32(node) {
			m.nodes[node.Service()] = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	m.Unlock()
}
