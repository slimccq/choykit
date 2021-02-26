// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//+build !ignore

package uuid

import (
	"sync"
	"testing"
	"time"
)

var (
	redisAddr = "localhost:6379"
	redisKey  = "uuid:counter"
	rkeys     = make(map[int64]bool)
	rguard    sync.Mutex
)

func TestRedisStoreExample(t *testing.T) {
	var store = NewRedisStore(redisAddr, redisKey)
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var start = time.Now()
	var count = 10000
	var ids []int64
	for i := 0; i < count; i++ {
		id := store.MustNext()
		ids = append(ids, id)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("QPS %.2f", float64(count)/elapsed)
	// Output: QPS 9393.92
}

// N个并发worker，测试生成id的一致性
func TestRedisStoreConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go etcdWorker(t, i, &wg)
	}
	wg.Wait()
}

func redisWorker(t *testing.T, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	var store = NewRedisStore(redisAddr, redisKey)
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var count = 1000
	for i := 0; i < count; i++ {
		id := store.MustNext()
		eguard.Lock()
		if _, found := ekeys[id]; found {
			eguard.Unlock()
			t.Fatalf("key %d exist", id)
			return
		}
		ekeys[id] = true
		eguard.Unlock()
	}
}
