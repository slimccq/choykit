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
	"errors"
	"sync/atomic"

	"devpkg.work/choykit/pkg/log"
)

const (
	StatCommit int = iota
	StatTimer
	StatExec
	StatError
	StatDropped
	NumStats
)

var (
	ErrExecutorClosed = errors.New("executor is closed")
	ErrExecutorBusy   = errors.New("executor is busy")
)

// Runner执行器
type Executor struct {
	Scheduler
	closing     int32               //
	concurrency int32               // 并发
	bus         chan Runner // 待执行队列
	stats       *Stats      // 执行统计
}

func (e *Executor) Init(queueSize, concurrency int) error {
	e.bus = make(chan Runner, queueSize)
	e.concurrency = int32(concurrency)
	e.stats = NewStats(NumStats)
	e.Scheduler.Init()
	return nil
}

func (e *Executor) Stats() *Stats {
	return e.stats.Clone()
}

func (e *Executor) Go() {
	e.Scheduler.Go()
	if e.concurrency <= 1 {
		e.wg.Add(1)
		go e.serveOne()
		return
	}
	// 多线程模式
	e.wg.Add(1)
	go e.serveTimer() // 所有的timer由同一个goroutine执行
	for i := 1; i < int(e.concurrency); i++ {
		e.wg.Add(1)
		go e.serveRunner(i)
	}
}

//
func (e *Executor) Shutdown() {
	log.Debugf("start shutdown executor")
	if atomic.CompareAndSwapInt32(&e.closing, 0, 1) {
		return
	}

	e.Scheduler.Shutdown()
	close(e.bus)
	e.bus = nil
	e.stats = nil
}

func (e *Executor) Execute(r Runner) error {
	if atomic.LoadInt32(&e.closing) == 1 {
		return ErrExecutorClosed
	}
	e.stats.Add(StatCommit, 1)
	e.showPending()
	e.bus <- r // may block current goroutine
	return nil
}

// 繁忙度
func (e *Executor) Busyness() float32 {
	return float32(len(e.bus)) / float32(cap(e.bus))
}

func (e *Executor) showPending() {
	if n := len(e.bus); n > cap(e.bus)/2 {
		log.Warnf("more than half runner(%d/%d) are pending!", n, len(e.bus)/2)
	}
}

// 执行剩下的runner
func (e *Executor) finally() {
	for r := range e.bus {
		e.run(r)
	}
}

func (e *Executor) run(r Runner) {
	defer Catch()
	if err := r.Run(); err != nil {
		e.stats.Add(StatError, 1)
		log.Errorf("execute runner (%T): %v", r, err)
	}
}

func (e *Executor) serveOne() {
	defer e.wg.Done()
	defer log.Debugf("executor stop serving with #%d runner left", len(e.bus))
	log.Debugf("executor start serving with capacity %d", cap(e.bus))
	for {
		var runner Runner
		select {
		case r, ok := <-e.bus:
			if !ok { // runner channel is closed
				return
			}
			runner = r

		case timer, ok := <-e.C:
			if !ok { // timer channel is closed
				return
			}
			runner = timer.R
			e.stats.Add(StatTimer, 1)

		case <-e.done:
			e.finally()
			log.Debugf("executor stop serving")
			return
		}

		e.showPending()
		e.run(runner)
		e.stats.Add(StatExec, 1)
	}
}

func (e *Executor) serveRunner(i int) {
	defer e.wg.Done()
	defer log.Debugf("executor stop serving runner")
	log.Debugf("executor start serving runner")
	defer e.finally()
	for {
		select {
		case r, ok := <-e.bus:
			if !ok { // runner channel is closed
				return
			}
			e.showPending()
			e.run(r)
			e.stats.Add(StatExec, 1)

		case <-e.done:
			return
		}
	}
}

func (e *Executor) serveTimer() {
	defer e.wg.Done()
	defer log.Debugf("executor stop serving timer")
	log.Debugf("executor start serving timer")
	for {
		select {
		case timer, ok := <-e.C:
			if !ok { // timer channel is closed
				return
			}
			e.showPending()
			e.stats.Add(StatTimer, 1)
			e.run(timer.R)
			e.stats.Add(StatExec, 1)

		case <-e.done:
			return
		}
	}
}
