// Copyright © 2021-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package cluster

import (
	"fmt"
	"testing"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/protocol"
)

type fakeServiceSink struct {
	nodes NodeInfoMap
}

func (s *fakeServiceSink) NodeInfo() *protocol.NodeInfo {
	return &protocol.NodeInfo{
		Node:      1234,
		Interface: "127.0.0.1:4321",
	}
}

func (s *fakeServiceSink) AddDependency(info *protocol.NodeInfo) {
	s.nodes.AddNode(info)
}

func (s *fakeServiceSink) DelDependency(removeAll bool, node choykit.NodeID) {
	if removeAll {
		s.nodes.Clear()
		return
	}
	s.nodes.DeleteNode(node)
}

func TestDiscoveryEtcd(t *testing.T) {
	opts := choykit.NewOptions()
	opts.EtcdAddress = "127.0.0.1:2379"
	opts.EtcdKeySpace = "/choyd"
	opts.EtcdLeaseTTL = 5
	sink := &fakeServiceSink{}
	d := NewEtcdDiscovery(opts, sink)
	defer d.Close()
	d.Start()
	<-d.done
	fmt.Printf("passed\n")
}
