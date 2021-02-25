// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
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
	"sync"
	"testing"
)

type testRunner2 struct {
	guard sync.Mutex
	count int32
	done  chan struct{}
}

func newtestRunner2() *testRunner2 {
	return &testRunner2{
		done: make(chan struct{}),
	}
}

func (r *testRunner2) Run() error {
	r.guard.Lock()
	r.count++
	if r.count == 11 {
		r.done <- struct{}{}
	}
	r.guard.Unlock()
	return nil
}

func TestExecutorExample(t *testing.T) {
	var exe Executor
	exe.Init(1024, 1)
	exe.Go()

	var r = newtestRunner2()
	exe.RunAfter(100, r)

	for i := 0; i < 10; i++ {
		exe.Execute(r)
	}
	<-r.done
	exe.Shutdown()
}

func TestExecutorExampleConcurrent(t *testing.T) {
	var exe Executor
	exe.Init(1024, 4)
	exe.Go()

	var r = newtestRunner2()
	exe.RunAfter(100, r)
	for i := 0; i < 10; i++ {
		exe.Execute(r)
	}
	<-r.done
	exe.Shutdown()
}
