// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"container/heap"
)

type TimerNode struct {
	ExpireTs int64  // 超时时间
	R        Runner // 超时后执行的runner
	Index    int32  // 数组索引
	id       int32  // 定时器ID
	interval int32  // 间隔（毫秒）
	repeat   int16  // 重复执行次数，负数表示一直执行
}

type TimerHeap []*TimerNode

func (q TimerHeap) Len() int {
	return len(q)
}

func (q TimerHeap) Less(i, j int) bool {
	return q[i].ExpireTs < q[j].ExpireTs
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

func (q *TimerHeap) Update(item *TimerNode, ts int64) {
	item.ExpireTs = ts
	heap.Fix(q, int(item.Index))
}
