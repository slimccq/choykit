// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "testing"

func TestPacketQueue(t *testing.T) {
	q := NewPacketQueue()
	for i := 0; i < 1000; i++ {
		pkt := MakePacket()
		q.Push(pkt)
	}
	for i := 0; i < 1000; i++ {
		q.Pop()
	}
}
