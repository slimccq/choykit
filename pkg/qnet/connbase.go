// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"net"
	"sync"
	"sync/atomic"

	"devpkg.work/choykit/pkg/fatchoy"
)

type ConnBase struct {
	done     chan struct{}           // done signal
	wg       sync.WaitGroup          // wait group
	closing  int32                   // closing flag
	node     fatchoy.NodeID          // node id
	addr     string                  // remote address
	userdata interface{}             // user data
	ctx      *fatchoy.ServiceContext // current service context object
	codec    fatchoy.Codec           // message encoding/decoding
	inbound  chan<- *fatchoy.Packet  // inbound message queue
	outbound chan *fatchoy.Packet    // outbound message queue
	stats    *fatchoy.Stats          // message stats
	errChan  chan error              // error signal
}

func (c *ConnBase) init(node fatchoy.NodeID, cdec fatchoy.Codec, inbound chan<- *fatchoy.Packet, outsize int,
	errChan chan error, stats *fatchoy.Stats) {
	if stats == nil {
		stats = fatchoy.NewStats(NumStat)
	}
	c.node = node
	c.codec = cdec
	c.stats = stats
	c.inbound = inbound
	c.errChan = errChan
	c.done = make(chan struct{})
	c.outbound = make(chan *fatchoy.Packet, outsize)
}

func (c *ConnBase) NodeID() fatchoy.NodeID {
	return c.node
}

func (c *ConnBase) SetNodeID(node fatchoy.NodeID) {
	c.node = node
}

func (c *ConnBase) SetRemoteAddr(addr string) {
	c.addr = addr
}

func (c *ConnBase) RemoteAddr() string {
	return c.addr
}

func (c *ConnBase) Stats() *fatchoy.Stats {
	return c.stats
}

func (c *ConnBase) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) == 1
}

func (c *ConnBase) Codec() fatchoy.Codec {
	return c.codec
}

func (c *ConnBase) Context() *fatchoy.ServiceContext {
	return c.ctx
}

func (c *ConnBase) SetContext(v *fatchoy.ServiceContext) {
	c.ctx = v
}

func (c *ConnBase) SetUserData(ud interface{}) {
	c.userdata = ud
}

func (c *ConnBase) UserData() interface{} {
	return c.userdata
}

// a fake endpoint
type FakeConn struct {
	ConnBase
}

func NewFakeConn(node fatchoy.NodeID, addr string) fatchoy.Endpoint {
	return &FakeConn{
		ConnBase: ConnBase{
			node: node,
			addr: addr,
		},
	}
}

func (c *FakeConn) RawConn() net.Conn {
	return nil
}

func (c *FakeConn) SendPacket(*fatchoy.Packet) error {
	return nil
}

func (c *FakeConn) Go(bool, bool) {
}

func (c *FakeConn) Close() error {
	return nil
}

func (c *FakeConn) ForceClose(error) {
}
