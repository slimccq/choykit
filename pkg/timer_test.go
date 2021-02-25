// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package choykit

import (
	"container/heap"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func verifyHeap(t *testing.T, h TimerHeap, i int) {
	n := h.Len()
	left := 2*i + 1
	right := 2*i + 2
	if left < n {
		if h.Less(left, i) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %v", i, h[i], left, h[right])
			return
		}
		verifyHeap(t, h, left)
	}
	if right < n {
		if h.Less(right, i) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %v", i, h[i], left, h[right])
			return
		}
		verifyHeap(t, h, right)
	}
}

func TestTimerHeap(t *testing.T) {
	var pq TimerHeap
	verifyHeap(t, pq, 0)

	if !pq.Empty() {
		t.Fatalf("pq not empty")
	}
	if pq.Peek() != nil {
		t.Fatalf("pq peek not nil")
	}

	var now = time.Now()
	for i := int32(1); i <= 100; i++ {
		var delay = rand.Int() % 1000
		var expire = now.Add(time.Millisecond * time.Duration(delay))
		item := &TimerNode{
			Priority: expire.UnixNano(),
			delay:    int32(delay),
			id:       i,
		}
		pq.Push(item)
	}

	heap.Init(&pq)
	verifyHeap(t, pq, 0)

	for i := int32(101); i <= 200; i++ {
		var delay = rand.Int() % 1000
		var expire = now.Add(time.Millisecond + time.Duration(delay))
		item := &TimerNode{
			Priority: expire.UnixNano(),
			delay:    int32(delay),
			id:       i,
		}
		heap.Push(&pq, item)
		verifyHeap(t, pq, 0)
	}

	if pq.Peek() == nil {
		t.Fatalf("pq peek should not be nil")
	}

	for pq.Len() > 0 {
		i := pq.Len() - 1
		item := heap.Remove(&pq, i).(*TimerNode)
		if item == nil {
			t.Fatalf("remove failed")
		}
		verifyHeap(t, pq, 0)
	}
}
