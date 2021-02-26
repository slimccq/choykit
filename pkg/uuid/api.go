// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
)

// 分布式UUID
var (
	seqIDGen      *SequenceID // 发号器算法
	uniqueGen     *SnowFlake  // 雪花算法
	useRedisStore = false
)

func Init(workerId uint16, addr, key string) {
	var store Storage
	if useRedisStore {
		store = NewRedisStore(addr, key)
	} else {
		store = NewEtcdStore(addr, key)
	}
	if err := store.Init(); err != nil {
		log.Fatalf("Storage.Init: %v", err)
	}
	var seq = NewSequenceID(store, DefaultSeqStep)
	if err := seq.Init(); err != nil {
		log.Fatalf("SequenceID.Init: %v", err)
	}
	seqIDGen = seq
	uniqueGen = NewSnowFlake(workerId)
}

// 生成依赖存储，可用于角色ID
func NextID() int64 {
	return seqIDGen.MustNext()
}

// 生成的值与时钟有关，通常值比较大，可用于日志ID
func NextUUID() int64 {
	return uniqueGen.Next()
}

func MustCreateGUID() string {
	u, err := NewV4()
	if err != nil {
		log.Panicf("uuid.NewV4: %v", err)
	}
	return u.String()
}
