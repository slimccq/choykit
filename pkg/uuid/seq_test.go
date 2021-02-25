// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"testing"
	"time"
)

const tetLoad = 2000000

func TestSequenceIDQPSEtcd(t *testing.T) {
	var store = NewEtcdStore(etcdAddr, "uuid:ctr123")
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var seq = NewSequenceID(store, 1000)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var m = make(map[int64]bool)
	var start = time.Now()
	for i := 0; i < tetLoad; i++ {
		uid := seq.Next()
		if _, found := m[uid]; found {
			t.Fatalf("key %d exist", uid)
		}
		m[uid] = true
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("SeqID QPS %f", float64(tetLoad)/elapsed)
}

func TestSequenceIDQPSRedis(t *testing.T) {
	var store = NewRedisStore(redisAddr, "uuid:ctr123")
	if err := store.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	var seq = NewSequenceID(store, 1000)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var m = make(map[int64]bool)
	var start = time.Now()
	for i := 0; i < tetLoad; i++ {
		uid := seq.Next()
		if _, found := m[uid]; found {
			t.Fatalf("key %d exist", uid)
		}
		m[uid] = true
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("SeqID QPS %f", float64(tetLoad)/elapsed)
}
