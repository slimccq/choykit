// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"devpkg.work/choykit/pkg/x/cipher"
	"net"
	"sync"
)

type MessageEndpoint interface {
	NodeID() NodeID
	SetNodeID(NodeID)
	RemoteAddr() string

	// 发送消息
	SendPacket(*Packet) error

	// 关闭读/写
	Close() error
	ForceClose(error)
	IsClosing() bool

	Context() *ServiceContext
	SetContext(*ServiceContext)
}

// 网络连接端点
type Endpoint interface {
	MessageEndpoint

	RawConn() net.Conn
	Stats() *Stats
	Encoder() ProtocolCodec

	Go(write, read bool)

	SetEncrypt(cipher.BlockCryptor, cipher.BlockCryptor)

	SetUserData(interface{})
	UserData() interface{}
}

// 线程安全的endpoint map
type EndpointMap struct {
	sync.RWMutex
	endpoints map[NodeID]Endpoint
}

func NewEndpointMap() *EndpointMap {
	return &EndpointMap{
		endpoints: make(map[NodeID]Endpoint, 8),
	}
}

func (e *EndpointMap) Get(node NodeID) Endpoint {
	e.RLock()
	v := e.endpoints[node]
	e.RUnlock()
	return v
}

func (e *EndpointMap) Add(node NodeID, endpoint Endpoint) {
	e.Lock()
	e.endpoints[node] = endpoint
	e.Unlock()
}

func (e *EndpointMap) Delete(node NodeID) bool {
	e.Lock()
	delete(e.endpoints, node)
	e.Unlock()
	return false
}

func (e *EndpointMap) Size() int {
	e.Lock()
	n := len(e.endpoints)
	e.Unlock()
	return n
}

func (e *EndpointMap) Reset() {
	e.Lock()
	e.endpoints = make(map[NodeID]Endpoint, 8)
	e.Unlock()
}

func (e *EndpointMap) List() []Endpoint {
	e.RLock()
	var endpoints = make([]Endpoint, 0, len(e.endpoints))
	for _, v := range e.endpoints {
		endpoints = append(endpoints, v)
	}
	e.RUnlock()
	return endpoints
}
