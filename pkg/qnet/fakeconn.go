// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"devpkg.work/choykit/pkg/fatchoy"
	"net"
)

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
