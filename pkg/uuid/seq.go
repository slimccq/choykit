// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
	"sync"
)

const DefaultSeqStep = 2000

type SequenceID struct {
	guard    sync.Mutex
	store    Storage // storage implementation
	rangeEnd int64   // range counter
	step     int64   // scope of range
	lastID   int64   // last generated ID
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

func (s *SequenceID) Init() (err error) {
	ctr, err := s.store.Next()
	if err != nil {
		return err
	}
	s.lastID = ctr * s.step
	s.rangeEnd = (ctr + 1) * s.step
	if s.rangeEnd < s.lastID {
		log.Panicf("SeqID gone backwards: %d -> %d", s.lastID, s.rangeEnd)
	}
	return nil
}

func (s *SequenceID) Next() int64 {
	s.guard.Lock()
	defer s.guard.Unlock()
	var next = s.lastID + 1
	if next <= s.rangeEnd {
		s.lastID = next
	} else {
		var ctr = s.store.MustNext()
		s.lastID = ctr * s.step
		s.rangeEnd = (ctr + 1) * s.step
		if s.rangeEnd < s.lastID {
			log.Panicf("SeqID gone backwards: %d -> %d", s.lastID, s.rangeEnd)
		}
		next = s.lastID + 1
		s.lastID = next
	}
	return next
}
