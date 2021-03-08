// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
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
	encoder  fatchoy.ProtocolCodec   // message encoding/decoding
	inbound  chan<- *fatchoy.Packet  // inbound message queue
	outbound chan *fatchoy.Packet    // outbound message queue
	stats    *fatchoy.Stats          // message stats
	errChan  chan error              // error signal
}

func (c *ConnBase) init(node fatchoy.NodeID, encoder fatchoy.ProtocolCodec, inbound chan<- *fatchoy.Packet, outsize int32,
	errChan chan error, stats *fatchoy.Stats) {
	if stats == nil {
		stats = fatchoy.NewStats(NumStat)
	}
	c.node = node
	c.encoder = encoder
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

func (c *ConnBase) Encoder() fatchoy.ProtocolCodec {
	return c.encoder
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

