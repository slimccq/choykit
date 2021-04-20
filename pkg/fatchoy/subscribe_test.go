// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "testing"

func TestMessageSubAdd(t *testing.T) {
	sub := NewMessageSub()
	sub.AddSubNodeRange(100, 200, 0x301)
	sub.AddSubNode(101, 0x302)
	nodes := sub.GetSubscribeNodes(100, 200)
	t.Logf("sub nodes: %v", nodes)
	nodes = sub.GetSubscribeNodesOf(101)
	t.Logf("sub nodes: %v", nodes)
}