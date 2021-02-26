// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "testing"

func TestNewPacket(t *testing.T) {
	pkt := NewPacket(1234, 1001, 2001, 1, 12, "hello")
	t.Logf("%v", pkt)
}