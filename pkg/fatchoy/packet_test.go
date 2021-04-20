// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "testing"

func TestNewPacket(t *testing.T) {
	pkt := NewPacket(1234, 1001,   0x1, "hello")
	t.Logf("new: %v", pkt)
	clone := pkt.Clone()
	pkt.Reset()
	t.Logf("reset: %v", pkt)
	t.Logf("clone: %v", clone)

	clone.SetErrno(1002)
	t.Logf("clone: %v", clone)
}