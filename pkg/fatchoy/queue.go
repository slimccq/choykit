// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"sync"

	"devpkg.work/choykit/pkg/x/collections"
)

// 一个无边界限制的Packet队列
type PacketQueue struct {
	mu  sync.Mutex        // mutex object
	rep collections.Queue // queue representation
	C   chan struct{}     // notify channel
}

func NewPacketQueue() *PacketQueue {
	q := &PacketQueue{}
	q.Reset()
	return q
}

func (q *PacketQueue) Reset() {
	q.mu.Lock()
	q.C = make(chan struct{}, 1)
	q.rep.Init()
	q.mu.Unlock()
}

func (q *PacketQueue) Notify() {
	select {
	case q.C <- struct{}{}:
		return
	default:
		return
	}
}

func (q *PacketQueue) Len() int {
	q.mu.Lock()
	n := q.rep.Len()
	q.mu.Unlock()
	return n
}

// 压入队列
func (q *PacketQueue) Push(v *Packet) {
	q.mu.Lock()
	q.rep.Push(v)
	q.mu.Unlock()
	q.Notify()
}

// 取出队列头部元素
func (q *PacketQueue) Peek() *Packet {
	q.mu.Lock()
	if v, ok := q.rep.Front(); !ok {
		q.mu.Unlock()
		return nil
	} else {
		m := v.(*Packet)
		q.mu.Unlock()
		return m
	}
}

// 弹出队列
func (q *PacketQueue) Pop() *Packet {
	q.mu.Lock()
	if v, ok := q.rep.Pop(); !ok {
		q.mu.Unlock()
		return nil
	} else {
		m := v.(*Packet)
		q.mu.Unlock()
		return m
	}
}
