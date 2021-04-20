// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"devpkg.work/choykit/pkg/x/cipher"
	"io"
)

// 消息编解码，同样一个codec会在多个goroutine执行，需要多线程安全
type ProtocolCodec interface {
	ProtocolEncoder
	ProtocolDecoder
}

// 协议编码实现
type ProtocolEncoder interface {
	// 把packet写入`w`，使用encrypt加密，返回写入字节和错误
	Marshal(w io.Writer, encrypt cipher.BlockCryptor, pkt *Packet) (int, error)
}

type ProtocolDecoder interface {
	// 从`r`中解码packet，使用decrypt解密，返回读取字节和错误
	Unmarshal(r io.Reader, decrypt cipher.BlockCryptor, pkt *Packet) (int, error)
}
