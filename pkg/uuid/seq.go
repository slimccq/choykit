// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"fmt"
	"log"
	"sync"
)

// 默认每个counter分配2000个号
const DefaultSeqStep = 2000

// 发号器
type SequenceID struct {
	guard    sync.Mutex
	store    Storage // 存储组件
	rangeEnd int64   // 当前号段最大值
	step     int64   // 号段区间范围
	lastID   int64   // 上次生成的ID
}

func NewSequenceID(store Storage, step int32) *SequenceID {
	if step <= 0 {
		step = DefaultSeqStep
	}
	return &SequenceID{
		store: store,
		step:  int64(step),
	}
}

func (s *SequenceID) Init() error {
	if err := s.reload(); err != nil {
		return err
	}
	if s.lastID + 10 >= s.rangeEnd {
		if err := s.reload(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SequenceID) reload() error {
	ctr, err := s.store.Next()
	if err != nil {
		return err
	}
	s.lastID = ctr * s.step
	s.rangeEnd = (ctr + 1) * s.step
	if s.rangeEnd < s.lastID {
		return fmt.Errorf("SeqID gone backwards: %d -> %d", s.lastID, s.rangeEnd)
	}
	return nil
}

func (s *SequenceID) Next() (int64, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	var next = s.lastID + 1
	if next <= s.rangeEnd {
		s.lastID = next
		return next, nil
	}
	if err := s.reload(); err != nil {
		return 0, err
	}
	next = s.lastID + 1
	s.lastID = next
	return next, nil
}

func (s *SequenceID) MustNext() int64 {
	n, err := s.Next()
	if err != nil {
		log.Panicf("%v", err)
	}
	return n
}
