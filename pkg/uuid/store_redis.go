// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisStore struct {
	addr string     // redis server address
	key  string     // name of key
	conn redis.Conn // connection
}

func NewRedisStore(addr, key string) Storage {
	return &RedisStore{
		addr: addr,
		key:  key,
	}
}

func (s *RedisStore) Init() error {
	if err := s.createConn(10); err != nil {
		return err
	}
	return nil
}

func (s *RedisStore) createConn(timeout int32) error {
	conn, err := redis.Dial("tcp", s.addr,
		redis.DialConnectTimeout(time.Duration(timeout)*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(5*time.Second))
	if err != nil {
		return err
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.conn = conn
	return nil
}

func (s *RedisStore) Next() (int64, error) {
	var err error
	var counter int64
	for i := 1; i <= 3; i++ {
		counter, err = redis.Int64(s.conn.Do("INCR", s.key))
		if err == nil {
			return counter, nil
		}
		if op, ok := err.(*net.OpError); ok { // try reconnect
			if op.Op == "write" || op.Op == "read" {
				var retry = i + 1
				var timeout = int32(retry*retry/3 + 1)
				if er := s.createConn(timeout); er != nil {
					return 0, err
				}
				continue
			}
		}
		return 0, err
	}
	return 0, err
}

func (s *RedisStore) MustNext() int64 {
	if counter, err := s.Next(); err != nil {
		log.Panicf("RedisStore.Next: %v", err)
		return 0
	} else {
		return counter
	}
}

func (s *RedisStore) Close() {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
