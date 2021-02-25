// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package choykit

import (
	"testing"
)

func TestNodeIDSimple(t *testing.T) {
	var node = NodeTypeClient
	if !node.IsClient() {
		t.Fatalf("node should be a client session")
	}
	node = MakeNodeID(0x3, 0xc, 0xd)
	t.Logf("node value: %v\n", node)
	if !node.IsBackend() {
		t.Fatalf("node should be a backend instance")
	}

	if node.District() != 0x3 {
		t.Fatalf("expect node district %x, but got %x", 0x3, node.District())
	}
	if node.Service() != 0xc {
		t.Fatalf("expect node service %x, but got %x", 0xc, node.Service())
	}
	if node.Instance() != 0xd {
		t.Fatalf("expect node instance %x, but got %x", 0xd, node.Instance())
	}
	node.SetDistrict(0xab)
	if node.District() != 0xab {
		t.Fatalf("expect node district %d, but got %d", 0xff, node.District())
	}
	node.SetService(0xcd)
	if node.Service() != 0xcd {
		t.Fatalf("expect node service %d, but got %d", 0xcd, node.Service())
	}
	node.SetInstance(0xef)
	if node.Instance() != 0xef {
		t.Fatalf("expect node group %d, but got %d", 0xef, node.Instance())
	}
}

func TestNodeIDParse(t *testing.T) {
	var node = MakeNodeID(0xab, 0xcd, 0xef)
	var n = MustParseNodeID("abcdef")
	if n != node {
		t.Fatalf("node not equal, %v != %v", node, n)
	}
}
