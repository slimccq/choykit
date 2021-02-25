// Copyright © 2015-present prototyped.cn. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package choykit

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type testRunner struct {
	count    int
	lastTime time.Time
}

func (r *testRunner) Run() error {
	r.lastTime = time.Now()
	r.count++
	return nil
}

func TestScheduler_RunAfter(t *testing.T) {
	var sched Scheduler
	sched.Init()
	sched.Go()
	defer sched.Shutdown()

	var runer testRunner
	var interval int32 = 500 // 0.5s

	var start = time.Now()
	sched.RunAfter(interval, &runer)
	_, ok := <-sched.C
	if !ok {
		return
	}
	var fired = time.Now()
	if duration := fired.Sub(start); duration < time.Duration(interval)*time.Millisecond {
		t.Fatalf("invalid fire duration: %v != %v", duration, interval)
	}
}

func TestScheduler_RunEvery(t *testing.T) {
	var sched Scheduler
	sched.Init()
	sched.Go()
	defer sched.Shutdown()

	var runner testRunner
	var interval int32 = 100 // 0.1s
	sched.RunEvery(interval, &runner)

	for fired := range sched.C {
		fired.R.Run()
		if runner.count == 10 {
			break
		}
	}
}
