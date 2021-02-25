// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

import (
	"container/heap"
)

type TimerNode struct {
	Priority int64  // absolute expire time
	delay    int32  // expire time relate to now
	repeat   int32  // timer repeat count
	id       int32  // unique timer ID
	Index    int32  // array index
	R        Runner // timer expire callback function
}

type TimerHeap []*TimerNode

func (q TimerHeap) Len() int {
	return len(q)
}

func (q TimerHeap) Less(i, j int) bool {
	return q[i].Priority < q[j].Priority
}

func (q TimerHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Index = int32(i)
	q[j].Index = int32(j)
}

func (q *TimerHeap) Push(x interface{}) {
	v := x.(*TimerNode)
	v.Index = int32(len(*q))
	*q = append(*q, v)
}

func (q *TimerHeap) Pop() interface{} {
	old := *q
	n := len(old)
	if n > 0 {
		v := old[n-1]
		v.Index = -1 // for safety
		*q = old[:n-1]
		return v
	}
	return nil
}

func (q TimerHeap) Peek() *TimerNode {
	if len(q) > 0 {
		return q[0]
	}
	return nil
}

func (q TimerHeap) Empty() bool {
	return len(q) == 0
}

func (q *TimerHeap) Update(item *TimerNode, priority int64) {
	item.Priority = priority
	heap.Fix(q, int(item.Index))
}
