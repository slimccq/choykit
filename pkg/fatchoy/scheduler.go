// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"devpkg.work/choykit/pkg/log"
)

const (
	TimerPrecision    = 10   // 精度为10ms
	TimerChanCapacity = 1000 //
)

type Scheduler struct {
	done   chan struct{}        //
	wg     sync.WaitGroup       //
	ticker *time.Ticker         // 系统定时器ticker
	guard  sync.Mutex           // heap guard
	timers TimerHeap            // 定时器heap
	nextId int64                // time id生成
	refs   map[int64]*TimerNode // 对timer进行O(1)查找
	C      chan *TimerNode      // 到期的定时器
}

func (s *Scheduler) Init() error {
	s.nextId = 1
	s.done = make(chan struct{})
	s.ticker = time.NewTicker(TimerPrecision * time.Millisecond)
	s.timers = make(TimerHeap, 0, 64)
	s.refs = make(map[int64]*TimerNode, 64)
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
			var now = t.UnixNano() / 1e6 // ns to ms
			var maxId = s.nextId
			for s.timers.Len() > 0 {
				var node = s.timers.Peek()
				if now < node.ExpireTs {
					break // no timer expired
				}
				// make sure we don't process timer created by timer events
				if node.id > maxId {
					continue
				}
				if node.repeat < 0 || node.repeat > 1 {
					if node.repeat > 1 { // is infinite
						node.repeat -= 1
					}
					var expire = now + int64(node.interval)
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

func (s *Scheduler) schedule(interval, repeat int32, r Runner) int64 {
	s.guard.Lock()
	var now = currentMs()

	// 假设ID一直自增不会溢出
	var id = s.nextId
	s.nextId++

	var node = &TimerNode{
		ExpireTs: now + int64(interval),
		interval: interval,
		repeat:   repeat,
		id:       id,
		R:        r,
	}
	heap.Push(&s.timers, node)
	s.refs[id] = node
	s.guard.Unlock()
	return id
}

// 创建一个定时器，在`interval`毫秒后运行`r`
func (s *Scheduler) RunAfter(interval int32, r Runner) int64 {
	if interval < 0 {
		interval = 0
	}
	if n := len(s.refs); n >= math.MaxUint16 {
		log.Errorf("RunAfter: timer id exhausted, current %d", n)
		return -1
	}
	return s.schedule(interval, 0, r)
}

// 创建一个定时器，每隔`interval`毫秒运行一次`r`
func (s *Scheduler) RunEvery(interval int32, r Runner) int64 {
	if interval <= 0 {
		interval = 100
	}
	if n := len(s.refs); n >= math.MaxUint16 {
		log.Errorf("RunEvery: timer id exhausted, current %d", n)
		return -1
	}
	return s.schedule(interval, -1, r)
}

func (s *Scheduler) Cancel(id int64) bool {
	s.guard.Lock()
	defer s.guard.Unlock()
	if timer, found := s.refs[id]; found {
		delete(s.refs, id)
		heap.Remove(&s.timers, int(timer.Index))
		return true
	}
	return false
}
