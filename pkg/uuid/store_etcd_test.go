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
	etcdKey  = "/uuid/ctr"
	ekeys    = make(map[int64]bool)
	eguard   sync.Mutex
)

func TestEtcdStoreExample(t *testing.T) {
	var store = NewEtcdStore(etcdAddr, etcdKey)
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var start = time.Now()
	var count = 100
	var ids []int64
	for i := 0; i < count; i++ {
		id := store.MustNext()
		ids = append(ids, id)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("etcd QPS %f", float64(count)/elapsed)
}

func TestEtcdStoreConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go etcdWorker(t, i, &wg)
	}
	wg.Wait()
}

func etcdWorker(t *testing.T, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	var store = NewEtcdStore(etcdAddr, etcdKey)
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
