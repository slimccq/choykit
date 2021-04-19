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

func newTestV2Packet(bodyLen int) *fatchoy.Packet {
	var packet = fatchoy.MakePacket()
	packet.Flag = 0x02
	packet.Command = 1234
	packet.Seq = 2012
	if bodyLen > 0 {
		packet.Body = strutil.RandBytes(bodyLen)
	}
	return packet
}

func testServerCodec(t *testing.T, c fatchoy.ProtocolCodec, size int, msgSent *fatchoy.Packet) {
	var encoded bytes.Buffer
	if err := c.Marshal(&encoded, msgSent); err != nil {
		t.Fatalf("Encode with size %d: %v", size, err)
	}
	var msgRecv fatchoy.Packet
	if _, err := c.Unmarshal(&encoded, &msgRecv); err != nil {
		t.Fatalf("Decode with size %d: %v", size, err)
	}
	if !isEqualPacket(t, msgSent, &msgRecv) {
		t.Fatalf("Encode and Decode not equal: %d\n%v\n%v", size, msgSent, msgRecv)
	}
}

func TestV2CodecSimpleEncode(t *testing.T) {
	var cdec = NewServerProtocolCodec()
	var sizeList = []int{0, 404, 1012, 2014, 4018, 8012, 40487, 1024 * 31, MaxAllowedServerCodecPayloadSize - 100} //
	for _, n := range sizeList {
		var pkt = newTestV2Packet(n)
		testServerCodec(t, cdec, n, pkt)
	}
}

func BenchmarkV2ProtocolMarshal(b *testing.B) {
	b.StopTimer()
	var cdec = NewServerProtocolCodec()
	var size = 4096
	b.Logf("benchmark with message size %d\n", size)
	var msg = newTestV2Packet(int(size))
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
