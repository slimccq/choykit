// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"io"
)

// 消息加解密
type MessageEncryptor interface {
	Encrypt(src []byte) ([]byte, error)
	Decrypt(src []byte) ([]byte, error)
}

// 消息编解码，同样一个codec会在多个goroutine执行，需要多线程安全
type ProtocolCodec interface {
	ProtocolEncoder
	ProtocolDecoder
}

type ProtocolEncoder interface {
	Marshal(w io.Writer, pkt *Packet) error
}

type ProtocolDecoder interface {
	Unmarshal(r io.Reader, pkt *Packet) (int, error)
}
