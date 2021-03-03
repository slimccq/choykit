// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package codec

import (
	"bytes"
	"testing"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/x/strutil"
)

func newTestV1Packet(bodyLen int) *fatchoy.Packet {
	var packet = fatchoy.MakePacket()
	packet.Flag = 0x0
	packet.Command = 2134
	packet.Seq = 2002
	if bodyLen > 0 {
		packet.Body = strutil.RandBytes(bodyLen)
	}
	return packet
}

func testV1Codec(t *testing.T, c fatchoy.ProtocolCodec, size int, msgToSend *fatchoy.Packet) {
	var encoded bytes.Buffer
	if err := c.Marshal(&encoded, msgToSend); err != nil {
		t.Fatalf("Encode failure: size %d, %v", size, err)
	}
	var msgToRecv fatchoy.Packet
	if _, err := c.Unmarshal(&encoded, &msgToRecv); err != nil {
		t.Fatalf("Decode failure: size: %d, %v", size, err)
	}
	if !isEqualPacket(t, msgToSend, &msgToRecv) {
		t.Fatalf("Encode and Decode not equal: size: %d, %v\n%v", size, msgToSend, msgToRecv)
	}
}

func TestV1CodecEncode(t *testing.T) {
	var cdec = NewClientProtocolCodec()
	var sizeList = []int{0 /*101, 202, 303, 404, 505, 606, 1012, 2014, 4018, */, 8012, MaxAllowedV1RecvBytes - 100} //
	for _, n := range sizeList {
		var pkt = newTestV1Packet(n)
		testV1Codec(t, cdec, n, pkt)
	}
}

func BenchmarkV1ProtocolMarshal(b *testing.B) {
	b.StopTimer()
	var cdec = NewClientProtocolCodec()
	var size = 1000
	b.Logf("benchmark with message size %d\n", size)
	var msg = newTestV1Packet(int(size))
	b.StartTimer()
	var buf bytes.Buffer
	if err := cdec.Marshal(&buf, msg); err != nil {
		b.Logf("Encode: %v", err)
	}
	var msg2 fatchoy.Packet
	if _, err := cdec.Unmarshal(&buf, &msg2); err != nil {
		b.Logf("Decode: %v", err)
	}
}
