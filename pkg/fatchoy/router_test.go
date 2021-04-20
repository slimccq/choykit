// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import "testing"

func TestRouterRoutingEntry(t *testing.T) {
	rtable := NewRoutingTable()
	rtable.AddEntry(101, 102)
	rtable.AddEntry(103, 102)
	dest := rtable.GetEntry(101)
	t.Logf("get: %v", dest)
	rtable.DeleteEntry(101)
	list := rtable.EntryList()
	t.Logf("list: %v", list)
}
