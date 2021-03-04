// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"math"
	"sync"

	"devpkg.work/choykit/pkg/fatchoy"
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

// 添加一个节点
func (m *NodeInfoMap) AddNode(info *protocol.NodeInfo) {
	m.Lock()
	node := fatchoy.NodeID(info.Node)
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
func (m *NodeInfoMap) DeleteNode(node fatchoy.NodeID) {
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
