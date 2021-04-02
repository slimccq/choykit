// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

var (
	etcdAddr = "127.0.0.1:2379"
)

func TestEtcdStoreExample(t *testing.T) {
	var store = NewEtcdStore(etcdAddr, "/uuid/ctr001")
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var (
		count = 10000
		ids   []int64
		m     = make(map[int64]bool)
	)
	var start = time.Now()
	for i := 0; i < count; i++ {
		id := store.MustNext()
		if _, found := m[id]; found {
			t.Fatalf("duplicate id %d", id)
		}
		ids = append(ids, id)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//    QPS 2619.06/s
}

func createEtcdStore(key string, t *testing.T) IDGenerator {
	var store = NewEtcdStore(etcdAddr, key)
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	return store
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestEtcdStoreDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 1
		eachMax = 1000
		m       = make(map[int64]int, 10000)
	)
	var start = time.Now()
	for i := 1; i <= gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			return createEtcdStore("uuid:ctr003", t)
		}
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	var elapsed = time.Now().Sub(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 2552.59/s
}
