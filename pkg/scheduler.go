// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"devpkg.work/choykit/pkg/log"
)

const (
	TimerTickInterval = 10  // every 10ms a tick
	TimerChanCapacity = 256 //
)

type Scheduler struct {
	done   chan struct{}        //
	wg     sync.WaitGroup       //
	ticker *time.Ticker         // 系统定时器ticker
	guard  sync.Mutex           // heap guard
	timers TimerHeap            // 定时器heap
	id     int32                // time id生成
	refs   map[int32]*TimerNode // 对timer进行O(1)查找
	C      chan *TimerNode      // 到期的定时器
}

func (s *Scheduler) Init() error {
	s.id = 1000 // magic number
	s.done = make(chan struct{})
	s.ticker = time.NewTicker(TimerTickInterval * time.Millisecond)
	s.timers = make(TimerHeap, 0, 16)
	s.refs = make(map[int32]*TimerNode, 16)
	s.C = make(chan *TimerNode, TimerChanCapacity)
	return nil
}

func (s *Scheduler) Shutdown() {
	s.ticker.Stop()
	close(s.done)
	s.wg.Wait()
	close(s.C)
	s.C = nil
	s.done = nil
	s.ticker = nil
	s.refs = nil
	s.timers = nil
}

// 当前毫秒时间
func currentMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func (s *Scheduler) serve() {
	defer s.wg.Done()
	defer log.Debugf("scheduler stop serving")
	log.Debugf("scheduler start serving")
	var expires = make([]*TimerNode, 0)
	for {
		select {
		case t := <-s.ticker.C:
			s.guard.Lock()
			var now = t.UnixNano() / 1e6
			for s.timers.Len() > 0 {
				var node = s.timers.Peek()
				if now < node.Priority {
					break // no timer expired
				}
				if node.repeat < 0 || node.repeat > 1 {
					if node.repeat > 1 { // is infinite
						node.repeat -= 1
					}
					var expire = now + int64(node.delay)
					s.timers.Update(node, expire)
				} else {
					heap.Pop(&s.timers)
					delete(s.refs, node.id)
				}
				expires = append(expires, node)
			}
			s.guard.Unlock()

			for _, timer := range expires {
				s.C <- timer
			}
			expires = expires[:0]

		case <-s.done:
			return
		}
	}
}

func (s *Scheduler) Go() {
	s.wg.Add(1)
	go s.serve()
}

func (s *Scheduler) counter() int32 {
	for i := math.MaxUint16; i > 0; i-- {
		if s.id < 0 {
			s.id = 0
		}
		s.id++
		if _, found := s.refs[s.id]; found {
			continue
		}
		break
	}
	return s.id
}

func (s *Scheduler) schedule(delay, repeat int32, r Runner) int32 {
	s.guard.Lock()
	var now = currentMs()
	var id = s.counter()
	var node = &TimerNode{
		Priority: now + int64(delay),
		delay:    delay,
		repeat:   repeat,
		id:       id,
		R:        r,
	}
	heap.Push(&s.timers, node)
	s.refs[id] = node
	s.guard.Unlock()
	return id
}

// 创建一个定时器，在`delay`毫秒后运行`r`
func (s *Scheduler) RunAfter(delay int32, r Runner) int32 {
	if delay < 0 {
		delay = 0
	}
	if n := len(s.refs); n >= math.MaxUint16 {
		log.Errorf("RunAfter: timer id exhausted, current %d", n)
		return -1
	}
	return s.schedule(delay, 0, r)
}

// 创建一个定时器，每隔`delay`毫秒运行一次`r`
func (s *Scheduler) RunEvery(delay int32, r Runner) int32 {
	if delay < 0 {
		delay = 1
	}
	if n := len(s.refs); n >= math.MaxUint16 {
		log.Errorf("RunEvery: timer id exhausted, current %d", n)
		return -1
	}
	return s.schedule(delay, -1, r)
}

func (s *Scheduler) Cancel(id int32) bool {
	s.guard.Lock()
	defer s.guard.Unlock()
	if timer, found := s.refs[id]; found {
		delete(s.refs, id)
		heap.Remove(&s.timers, int(timer.Index))
		return true
	}
	return false
}
