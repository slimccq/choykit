// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"io"
)

type MessageEncryptor interface {
	Encrypt(src []byte) ([]byte, error)
	Decrypt(src []byte) ([]byte, error)
}

// 消息编解码
type ProtocolCodec interface {
	Marshal(w io.Writer, pkt *Packet) error
	Unmarshal(r io.Reader, pkt *Packet) (int, error)
}
