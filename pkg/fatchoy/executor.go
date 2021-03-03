// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"errors"
	"os"
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
)

// Runner执行器
type Executor struct {
	Scheduler
	closing     int32       //
	concurrency int32       // 并发数
	bus         chan Runner // 待执行runner队列
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
	e.start()
}

func (e *Executor) start() {
	if e.concurrency <= 1 {
		e.wg.Add(1)
		go e.serveAllInOne()
		return
	}
	// 多线程模式
	e.wg.Add(1)
	// 所有的timer由同一个goroutine执行
	go e.serveTimer()
	for i := 1; i <= int(e.concurrency); i++ {
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

// 超过1/2的runner在等待的时候打印警告
func (e *Executor) showPending() {
	half := cap(e.bus) / 2
	if n := len(e.bus); n > half {
		log.Warnf("more than half runners(%d/%d) are pending!!!", n, half)
	}
}

func (e *Executor) catch(r Runner) {
	if v := recover(); v != nil {
		e.stats.Add(StatError, 1)
		Backtrace(v, os.Stderr)
	}
}

func (e *Executor) run(r Runner) {
	defer e.catch(r)
	e.stats.Add(StatExec, 1)
	if err := r.Run(); err != nil {
		e.stats.Add(StatError, 1)
		log.Errorf("execute runner (%T): %v", r, err)
	}
}

func (e *Executor) serveAllInOne() {
	defer func() {
		e.wg.Done()
		e.stats.Add(StatDropped, int64(len(e.bus)))
		log.Debugf("executor stop serving, #%d runner left", len(e.bus))
	}()
	log.Debugf("executor start serving, capacity %d", cap(e.bus))
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
			return
		}

		e.showPending()
		e.run(runner)
	}
}

func (e *Executor) serveRunner(idx int) {
	defer func() {
		e.wg.Done()
		log.Debugf("executor #%d stop serving, #%d runner left", idx, len(e.bus))
	}()

	log.Debugf("executor #%d start serving runner", idx)
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
	defer func() {
		e.wg.Done()
		e.stats.Add(StatDropped, int64(len(e.bus)))
		log.Debugf("executor stop serving timer")
	}()

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

		case <-e.done:
			return
		}
	}
}
