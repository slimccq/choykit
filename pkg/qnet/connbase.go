// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"net"
	"sync"
	"sync/atomic"

	"devpkg.work/choykit/pkg"
)

type ConnBase struct {
	done     chan struct{}           // done signal
	wg       sync.WaitGroup          // wait group
	closing  int32                   // closing flag
	node     choykit.NodeID          // node id
	addr     string                  // remote address
	userdata interface{}             // user data
	ctx      *choykit.ServiceContext // current service context object
	codec    choykit.Codec           // message encoding/decoding
	inbound  chan<- *choykit.Packet  // inbound message queue
	outbound chan *choykit.Packet    // outbound message queue
	stats    *choykit.Stats          // message stats
	errChan  chan error              // error signal
}

func (c *ConnBase) init(node choykit.NodeID, cdec choykit.Codec, inbound chan<- *choykit.Packet, outsize int,
	errChan chan error, stats *choykit.Stats) {
	if stats == nil {
		stats = choykit.NewStats(NumStat)
	}
	c.node = node
	c.codec = cdec
	c.stats = stats
	c.inbound = inbound
	c.errChan = errChan
	c.done = make(chan struct{})
	c.outbound = make(chan *choykit.Packet, outsize)
}

func (c *ConnBase) NodeID() choykit.NodeID {
	return c.node
}

func (c *ConnBase) SetNodeID(node choykit.NodeID) {
	c.node = node
}

func (c *ConnBase) SetRemoteAddr(addr string) {
	c.addr = addr
}

func (c *ConnBase) RemoteAddr() string {
	return c.addr
}

func (c *ConnBase) Stats() *choykit.Stats {
	return c.stats
}

func (c *ConnBase) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) == 1
}

func (c *ConnBase) Codec() choykit.Codec {
	return c.codec
}

func (c *ConnBase) Context() *choykit.ServiceContext {
	return c.ctx
}

func (c *ConnBase) SetContext(v *choykit.ServiceContext) {
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

func NewFakeConn(node choykit.NodeID, addr string) choykit.Endpoint {
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

func (c *FakeConn) SendPacket(*choykit.Packet) error {
	return nil
}

func (c *FakeConn) Go(bool, bool) {
}

func (c *FakeConn) Close() error {
	return nil
}

func (c *FakeConn) ForceClose(error) {
}
