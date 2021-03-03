// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"io"
)

type MessageEncryptor interface {
	Encrypt(src []byte, dst []byte) error
	Decrypt(src []byte, dst []byte) error
}

// 消息编码器
type MessageCodec interface {
	Clone() MessageCodec

	// 消息编解码
	Decode(r io.Reader, pkt *Packet, encrypt MessageEncryptor) (int, error)
	Encode(pkt *Packet, w io.Writer, decrypt MessageEncryptor) error
}
