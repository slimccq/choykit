// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var ErrCannotPutEtcd = errors.New("cannot put counter to etcd")

// 使用etcd的key的版本号自增实现
type EtcdStore struct {
	addr string // etcd地址
	key  string //
}

func NewEtcdStore(addr, key string) Storage {
	return &EtcdStore{
		addr: addr,
		key:  key,
	}
}

func (s *EtcdStore) Init() error {
	return nil
}

func (s *EtcdStore) Close() {
}

func (s *EtcdStore) putKey(key string, response *PutResponse) error {
	// etcd v3.3 uses http://host/v3beta/*
	// etcd v3.4 uses http://host/v3/*
	url := fmt.Sprintf("http://%s/v3/kv/put", s.addr)
	data, err := json.Marshal(map[string]interface{}{
		"key":    base64.StdEncoding.EncodeToString([]byte(key)),
		"value":  base64.StdEncoding.EncodeToString([]byte(key)),
		"prevKv": true,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	rawbytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, rawbytes)
	}
	// println(string(rawbytes))
	if err := json.Unmarshal(rawbytes, response); err != nil {
		return err
	}
	return nil
}

func (s *EtcdStore) Next() (int64, error) {
	var resp PutResponse
	if err := s.putKey(s.key, &resp); err != nil {
		return 0, err
	}
	if resp.PrevKv == nil {
		return 1, nil
	}
	rev, _ := strconv.ParseInt(resp.PrevKv.Version, 10, 64)
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

// from go.etcd.io/etcd/mvcc/mvccpb

type PutResponse struct {
	Header *ResponseHeader `json:"header,omitempty"`
	// if prev_kv is set in the request, the previous key-value pair will be returned.
	PrevKv *KeyValue `json:"prev_kv,omitempty"`
}

type ResponseHeader struct {
	// cluster_id is the ID of the cluster which sent the response.
	ClusterId string `json:"cluster_id,omitempty"`
	// member_id is the ID of the member which sent the response.
	MemberId string `json:"member_id,omitempty"`
	// revision is the key-value store revision when the request was applied.
	// For watch progress responses, the header.revision indicates progress. All future events
	// recieved in this stream are guaranteed to have a higher revision number than the
	// header.revision number.
	Revision string `json:"revision,omitempty"`
	// raft_term is the raft term when the request was applied.
	RaftTerm string `json:"raft_term,omitempty"`
}

type KeyValue struct {
	// key is the key in bytes. An empty key is not allowed.
	Key string `json:"key,omitempty"`
	// create_revision is the revision of last creation on this key.
	CreateRevision string `json:"create_revision,omitempty"`
	// mod_revision is the revision of last modification on this key.
	ModRevision string `json:"mod_revision,omitempty"`
	// version is the version of the key. A deletion resets
	// the version to zero and any modification of the key
	// increases its version.
	Version string `json:"version,omitempty"`
	// value is the value held by the key, in bytes.
	Value string `json:"value,omitempty"`
	// lease is the ID of the lease that attached to key.
	// When the attached lease expires, the key will be deleted.
	// If lease is 0, then no lease is attached to the key.
	Lease string `json:"lease,omitempty"`
}
