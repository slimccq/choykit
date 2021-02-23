// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var ErrCannotPutEtcd = errors.New("cannot put counter to etcd")

type EtcdStore struct {
	addr string
	key  string
	cli  *clientv3.Client
}

func NewEtcdStore(addr, key string) Storage {
	return &EtcdStore{
		addr: addr,
		key:  key,
	}
}

func (s *EtcdStore) Init() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(s.addr, ","),
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return err
	}
	s.cli = cli
	return nil
}

func (s *EtcdStore) Close() {
	if s.cli != nil {
		s.cli.Close()
		s.cli = nil
	}
}

func (s *EtcdStore) Next() (int64, error) {
	var err error
	var now = time.Now().UnixNano()
	var value = strconv.FormatInt(now/1e6, 10) // ms
	resp, err := s.cli.Put(context.TODO(), s.key, value, clientv3.WithPrevKV())
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, ErrCannotPutEtcd
	}
	if resp.PrevKv == nil {
		return 1, nil
	}
	var cnt = resp.PrevKv.Version + 1
	return cnt, nil
}

func (s *EtcdStore) MustNext() int64 {
	if counter, err := s.Next(); err != nil {
		log.Panicf("EtcdStore.Next: %v", err)
		return 0
	} else {
		return counter
	}
}
