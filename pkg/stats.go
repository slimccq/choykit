// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import (
	"sync"
)

type Stats struct {
	mu  sync.RWMutex
	arr []int64
}

func NewStats(n int) *Stats {
	return &Stats{arr: make([]int64, n)}
}

func (s *Stats) Get(field int) (v int64) {
	s.mu.RLock()
	if field >= 0 && field < len(s.arr) {
		v = s.arr[field]
	}
	s.mu.RUnlock()
	return v
}

func (s *Stats) Set(field int, v int64) {
	s.mu.Lock()
	if field >= 0 && field < len(s.arr) {
		s.arr[field] = v
	}
	s.mu.Unlock()
}

func (s *Stats) Add(field int, delta int64) (v int64) {
	s.mu.Lock()
	if field >= 0 && field < len(s.arr) {
		s.arr[field] += delta
		v = s.arr[field]
	}
	s.mu.Unlock()
	return v
}

func (s *Stats) Values() []int64 {
	v := make([]int64, len(s.arr))
	s.mu.RLock()
	copy(v, s.arr)
	s.mu.RUnlock()
	return v
}

func (s *Stats) Clone() *Stats {
	return &Stats{arr: s.Values()}
}
