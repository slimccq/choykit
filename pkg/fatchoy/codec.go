// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

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
