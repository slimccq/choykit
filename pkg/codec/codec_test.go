// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package codec

import (
	"bytes"
	"testing"

	"devpkg.work/choykit/pkg"
)

func isEqualPacket(t *testing.T, a, b *choykit.Packet) bool {
	if a.Command != b.Command ||
		(a.Seq != b.Seq) ||
		(a.Flags != b.Flags) ||
		(a.Referer != b.Referer) ||
		(a.Node != b.Node) {
		return false
	}
	data1, _ := a.Encode()
	data2, _ := b.Encode()
	if len(data1) > 0 && len(data2) > 0 {
		if !bytes.Equal(data1, data2) {
			t.Fatalf("packet not equal, %v != %v", a, b)
			return false
		}
	}
	return true
}
