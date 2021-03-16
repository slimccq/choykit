// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"fmt"
	"log"
	"strconv"
)

const (
	NodeServiceShift = 16
	NodeServiceMask  = 0xFF00FFFF
	NodeInstanceMask = 0xFFFF0000
	NodeTypeShift    = 31
	NodeTypeClient   = NodeID(1 << NodeTypeShift)
)

// 一个32位整数表示的节点号，用以标识一个service（最高位为0），或者一个客户端session(最高位为1)
//
//	服务实例二进制布局
// 		--------------------------------------------
// 		| type | reserved |  service  |  instance  |
// 		--------------------------------------------
// 		32    31          28         16            0
//
//		8位服务编号，16位服务实例编号
//

// 节点号
type NodeID uint32

func MakeNodeID(service uint8, instance uint16) NodeID {
	return NodeID((uint32(service) << NodeServiceShift) | uint32(instance))
}

func MakeSessionNodeID(session uint32) NodeID {
	return NodeTypeClient | NodeID(session)
}

func MustParseNodeID(s string) NodeID {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Panicf("MustParseNode: %v", err)
	}
	return NodeID(n)
}

// 是否backend instance
func (n NodeID) IsBackend() bool {
	return (uint32(n) & uint32(NodeTypeClient)) == 0
}

// 服务类型编号
func (n NodeID) Service() uint8 {
	return uint8(n >> NodeServiceShift)
}

func (n *NodeID) SetService(v uint8) {
	var node = (uint32(*n) & NodeServiceMask) | (uint32(v) << NodeServiceShift)
	*n = NodeID(node)
}

// 实例编号
func (n NodeID) Instance() uint16 {
	return uint16(n)
}

func (n *NodeID) SetInstance(v uint16) {
	var node = (uint32(*n) & uint32(NodeInstanceMask)) | uint32(v)
	*n = NodeID(node)
}

func (n NodeID) String() string {
	return fmt.Sprintf("%02x%04x", n.Service(), n.Instance())
}

// 没有重复Node的集合
type NodeIDSet []NodeID

func (s NodeIDSet) Add(node NodeID) NodeIDSet {
	for _, n := range s {
		if n == node {
			return s
		}
	}
	return append(s, node)
}

func (s NodeIDSet) Has(node NodeID) int {
	for i, n := range s {
		if n == node {
			return i
		}
	}
	return -1
}

func (s NodeIDSet) Delete(node NodeID) NodeIDSet {
	for i, n := range s {
		if n == node {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (s NodeIDSet) Copy() []NodeID {
	if len(s) == 0 {
		return nil
	}
	v := make([]NodeID, len(s))
	copy(v, s)
	return v
}
