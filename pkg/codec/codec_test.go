// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package codec

import (
	"bytes"
	"testing"

	"devpkg.work/choykit/pkg/fatchoy"
)

func isEqualPacket(t *testing.T, a, b *fatchoy.Packet) bool {
	if a.Command != b.Command || (a.Seq != b.Seq) || (a.Flag != b.Flag) || (a.Node != b.Node) {
		return false
	}
	data1, _ := a.EncodeBody()
	data2, _ := b.EncodeBody()
	if len(data1) > 0 && len(data2) > 0 {
		if !bytes.Equal(data1, data2) {
			t.Fatalf("packet not equal, %v != %v", a, b)
			return false
		}
	}
	return true
}

func TestEncodePacket(t *testing.T) {
	pkt := newTestV1Packet(1024)
	if err := EncodePacket(pkt, 1000, nil); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if err := DecodePacket(pkt, nil); err != nil {
		t.Fatalf("decode: %v", err)
	}
}