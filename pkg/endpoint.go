// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

import (
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
	Codec() Codec

	Go(write, read bool)

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
