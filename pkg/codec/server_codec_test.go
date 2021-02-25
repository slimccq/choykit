// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package codec

import (
	"bytes"
	"testing"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/x/strutil"
)

func newTestV2Packet(bodyLen int) *choykit.Packet {
	var packet = choykit.MakePacket()
	packet.Flags = 0x0
	packet.Node = 0x0501
	packet.Command = 1234
	packet.Seq = 2012
	packet.Referer = 12345678
	if bodyLen > 0 {
		packet.Body = strutil.RandBytes(bodyLen)
	}
	return packet
}

func testServerCodec(t *testing.T, c choykit.Codec, size int, msgSent *choykit.Packet) {
	var encoded bytes.Buffer
	if err := c.Encode(msgSent, &encoded); err != nil {
		t.Fatalf("Encode with size %d: %v", size, err)
	}
	var msgRecv choykit.Packet
	if _, err := c.Decode(&encoded, &msgRecv); err != nil {
		t.Fatalf("Decode with size %d: %v", size, err)
	}
	if !isEqualPacket(t, msgSent, &msgRecv) {
		t.Fatalf("Encode and Decode not equal: %d\n%v\n%v", size, msgSent, msgRecv)
	}
}

func TestV2CodecSimpleEncode(t *testing.T) {
	var cdec = NewServerCodec()
	var sizeList = []int{0, 404, 1012, 2014, 4018, 8012, 40487, 1024 * 31, MaxAllowedV2PayloadSize - 100} //
	for _, n := range sizeList {
		var pkt = newTestV2Packet(n)
		testServerCodec(t, cdec, n, pkt)
	}
}

func BenchmarkV2ProtocolMarshal(b *testing.B) {
	b.StopTimer()
	var cdec = NewServerCodec()
	var size = 4096
	b.Logf("benchmark with message size %d\n", size)
	var msg = newTestV2Packet(int(size))
	b.StartTimer()

	var buf bytes.Buffer
	if err := cdec.Encode(msg, &buf); err != nil {
		b.Logf("Encode: %v", err)
	}
	var msg2 choykit.Packet
	if _, err := cdec.Decode(&buf, &msg2); err != nil {
		b.Logf("Decode: %v", err)
	}
}
