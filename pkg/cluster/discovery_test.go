// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cluster

import (
	"fmt"
	"testing"

	"devpkg.work/choykit/pkg/fatchoy"
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

func (s *fakeServiceSink) DelDependency(removeAll bool, node fatchoy.NodeID) {
	if removeAll {
		s.nodes.Clear()
		return
	}
	s.nodes.DeleteNode(node)
}

func TestDiscoveryEtcd(t *testing.T) {
	opts := fatchoy.NewOptions()
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
