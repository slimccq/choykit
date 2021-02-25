// Copyright © 2020-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

import (
	"sync"

	"devpkg.work/choykit/pkg/x/collections"
)

// A memory-bound packet queue
type PacketQueue struct {
	mu  sync.Mutex        // mutex object
	rep collections.Queue // queue representation
	C   chan struct{}     // notify channel
}

func NewMessageQueue() *PacketQueue {
	q := &PacketQueue{}
	q.Reset()
	return q
}

// initializes or clears
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

func (q *PacketQueue) Push(v *Packet) {
	q.mu.Lock()
	q.rep.Push(v)
	q.mu.Unlock()
	q.Notify()
}

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
