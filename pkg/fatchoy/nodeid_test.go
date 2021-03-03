// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package fatchoy

import (
	"fmt"
	"testing"
)

func TestNodeIDSimple(t *testing.T) {
	var node = NodeTypeClient
	if node.IsBackend() {
		t.Fatalf("node should be a client session")
	}
	var srv = uint8(0xef)
	var inst = uint16(0xabcd)
	node = MakeNodeID(srv, inst)
	t.Logf("node value: %v\n", node)
	if !node.IsBackend() {
		t.Fatalf("node should be a backend instance")
	}

	if v := node.Service(); v != srv {
		t.Fatalf("expect node service %x, but got %x", inst, v)
	}
	if v := node.Instance(); v != inst {
		t.Fatalf("expect node instance %x, but got %x", inst, node.Instance())
	}
	srv = uint8(0xab)
	inst = uint16(0xcdef)
	node.SetService(srv)
	if v := node.Service(); v != srv {
		t.Fatalf("expect node service %x, but got %x", srv, node.Service())
	}
	node.SetInstance(inst)
	if v := node.Instance(); v != inst {
		t.Fatalf("expect node group %x, but got %x", inst, node.Instance())
	}
}

func TestNodeIDParse(t *testing.T) {
	srv := uint8(0xab)
	inst := uint16(0xcdef)
	var node = MakeNodeID(srv, inst)
	var n = MustParseNodeID("abcdef")
	if n != node {
		t.Fatalf("node not equal, %v != %v", node, n)
	}
}

func TestNodeIDSet(t *testing.T) {
	var set NodeIDSet
	for i := 1; i <= 5; i++ {
		nid := NodeID(0x10000 + i)
		set = set.Add(nid)
		if set.Has(nid) < 0 {
			t.Fatalf("node %v check fail", nid)
		}
	}
	nid := NodeID(0x10001)
	set = set.Delete(nid)
	if set.Has(nid) >= 0 {
		t.Fatalf("node %v check fail", nid)
	}
	copy := set.Copy()
	fmt.Printf("origin: %v\n", set)
	fmt.Printf("copy  : %v\n", copy)
}
