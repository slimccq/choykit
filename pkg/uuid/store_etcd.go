// Copyright © 2020-present ichenq@outlook.com All rights reserved.
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

var (
	ErrCannotPutEtcd = errors.New("cannot put counter to etcd")
)

// 使用etcd的key的版本号自增实现
type EtcdStore struct {
	addrList []string // etcd地址
	key      string   //
	cli      *clientv3.Client
}

func NewEtcdStore(addr, key string) Storage {
	return &EtcdStore{
		addrList: strings.Split(addr, ","),
		key:      key,
	}
}

func (s *EtcdStore) Init() error {
	return s.createClient()
}

func (s *EtcdStore) Close() error {
	if s.cli != nil {
		s.cli.Close()
		s.cli = nil
	}
	return nil
}

func (s *EtcdStore) createClient() error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   s.addrList,
		DialTimeout: time.Second * TimeoutSec,
	})
	if err != nil {
		return err
	}
	s.cli = client
	return nil
}

func (s *EtcdStore) putKey() (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	value := strconv.FormatInt(time.Now().Unix(), 10)
	resp, err := s.cli.Put(ctx, s.key, value, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *EtcdStore) Next() (int64, error) {
	resp, err := s.putKey()
	if err != nil {
		return 0, err
	}
	if resp.PrevKv == nil {
		return 1, nil
	}
	rev := resp.PrevKv.Version
	return rev + 1, nil
}

func (s *EtcdStore) MustNext() int64 {
	if counter, err := s.Next(); err != nil {
		log.Panicf("EtcdStore.Next: %v", err)
		return 0
	} else {
		return counter
	}
}
