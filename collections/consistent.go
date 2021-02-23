// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"fmt"
	"hash/crc32"
	"sort"
)

const (
	ReplicaCount = 20 // 虚拟节点数量
)

// 一致性hash
type Consistent struct {
	circle     map[uint32]string // hash环
	nodes      map[string]bool   // 所有节点
	sortedHash []uint32          // 环hash排序
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle: make(map[uint32]string),
		nodes:  make(map[string]bool),
	}
}

func (c *Consistent) hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// 添加一个节点
func (c *Consistent) AddNode(node string) {
	for i := 0; i < ReplicaCount; i++ {
		var replica = fmt.Sprintf("%s-%d", node, i)
		c.circle[c.hashKey(replica)] = node
	}
	c.nodes[node] = true
	c.updateSortedHash()
}

func (c *Consistent) RemoveNode(node string) {
	for i := 0; i < ReplicaCount; i++ {
		var replica = fmt.Sprintf("%s-%d", node, i)
		var key = c.hashKey(replica)
		delete(c.circle, key)
	}
	delete(c.nodes, node)
	c.updateSortedHash()
}

// 获取一个节点
func (c *Consistent) GetNodeBy(key string) string {
	var i = c.search(c.hashKey(key))
	var node = c.circle[c.sortedHash[i]]
	return node
}

// 找到第一个大于等于`hash`的节点
func (c *Consistent) search(hash uint32) int {
	var lo = 0
	var hi = len(c.sortedHash)
	for lo < hi {
		var mid = lo + (hi-lo)/2
		if c.sortedHash[mid] <= hash {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= len(c.sortedHash) {
		lo = 0
	}
	return lo
}

func (c *Consistent) updateSortedHash() {
	hashes := c.sortedHash[:0]
	// 使用率低于1/4重新分配内存
	if cap(c.sortedHash)/(ReplicaCount*4) > len(c.circle) {
		hashes = nil
	}
	for k, _ := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(U32Array(hashes))
	c.sortedHash = hashes
}

type U32Array []uint32

func (x U32Array) Len() int           { return len(x) }
func (x U32Array) Less(i, j int) bool { return x[i] < x[j] }
func (x U32Array) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
