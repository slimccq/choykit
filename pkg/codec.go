// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

import (
	"bytes"
	"io"
)

// 消息编码器
type Codec interface {
	Version() uint8
	Clone() Codec

	SetSeqNo(seq uint16)
	SetEncryptKey(key, iv []byte)

	// 消息编解码
	Decode(rd io.Reader, pkt *Packet) (int, error)
	Encode(pkt *Packet, buf *bytes.Buffer) error
}
