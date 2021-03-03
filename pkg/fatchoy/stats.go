// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"sync/atomic"
)

// 一组计数器
type Stats struct {
	arr []int64
}

func NewStats(n int) *Stats {
	return &Stats{arr: make([]int64, n)}
}

func (s *Stats) Get(i int) int64 {
	if i >= 0 && i < len(s.arr) {
		return atomic.LoadInt64(&s.arr[i])
	}
	return 0
}

func (s *Stats) Set(i int, v int64) {
	if i >= 0 && i < len(s.arr) {
		atomic.StoreInt64(&s.arr[i], v)
	}
}

func (s *Stats) Add(i int, delta int64) int64 {
	if i >= 0 && i < len(s.arr) {
		return atomic.AddInt64(&s.arr[i], delta)
	}
	return 0
}

func (s *Stats) Copy() []int64 {
	arr := make([]int64, len(s.arr))
	for i := 0; i < len(arr); i++ {
		arr[i] = atomic.LoadInt64(&s.arr[i])
	}
	return arr
}

func (s *Stats) Clone() *Stats {
	return &Stats{arr: s.Copy()}
}
