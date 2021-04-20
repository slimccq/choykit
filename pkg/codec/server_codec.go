// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"devpkg.work/choykit/pkg/x/cipher"
	"encoding/binary"
	"hash/crc32"
	"io"

	"devpkg.work/choykit/pkg/fatchoy"
	"github.com/pkg/errors"
)

const (
	ServerCodecVersion    = 2  // 协议版本
	ServerCodecHeaderSize = 18 // 消息头大小
)

var MaxAllowedV2CodecPayloadSize = 8 * 1024 * 1024 // 最大包体大小(8M)

// wire format of codec header
//       --------------------------------
// field | len | flag | seq | cmd | crc |
//       --------------------------------
// bytes |  4  |  2   |  4  |  4  |  4  |
//       --------------------------------

// 编码器
type serverProtocolCodec struct {
}

var ServerProtocolCodec = NewServerProtocolCodec()

func NewServerProtocolCodec() fatchoy.ProtocolCodec {
	return &serverProtocolCodec{}
}

// 把消息内容写入w，并按需对body加密，消息压缩请在上层处理
func (c *serverProtocolCodec) Marshal(w io.Writer, encryptor cipher.BlockCryptor,pkt *fatchoy.Packet) (int, error) {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return 0, err
	}
	if len(payload) > 0 && encryptor != nil {
		payload = encryptor.Encrypt(payload)
		pkt.Flag |= fatchoy.PacketFlagEncrypted
	}
	var maxN = MaxAllowedV2CodecPayloadSize
	if n := len(payload); n >= maxN {
		return 0, errors.Errorf("message %d too large payload %d/%d", pkt.Command, n, maxN)
	}

	n := len(payload)
	hash := crc32.NewIEEE()
	var headbuf [ServerCodecHeaderSize]byte
	binary.LittleEndian.PutUint32(headbuf[0:], uint32(n))
	binary.LittleEndian.PutUint16(headbuf[4:], pkt.Flag)
	binary.LittleEndian.PutUint32(headbuf[6:], pkt.Seq)
	binary.LittleEndian.PutUint32(headbuf[10:], pkt.Command)
	hash.Write(headbuf[:ServerCodecHeaderSize-4])
	if n > 0 {
		hash.Write(payload)
	}
	binary.LittleEndian.PutUint32(headbuf[ServerCodecHeaderSize-4:], hash.Sum32())
	nbytes, err := w.Write(headbuf[0:])
	if err == nil && n > 0  {
		n, err = w.Write(payload)
		nbytes += n
	}
	return nbytes, err
}

// 从r中读取消息内容，检查包体大小和校验码，和解密，解压缩请在之后处理
func (c *serverProtocolCodec) Unmarshal(r io.Reader, decryptor cipher.BlockCryptor,pkt *fatchoy.Packet) (int, error) {
	var headbuf [ServerCodecHeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return 0, err
	}

	bodyLen := int(binary.LittleEndian.Uint32(headbuf[0:]))
	pkt.Flag = binary.LittleEndian.Uint16(headbuf[4:])
	pkt.Seq = binary.LittleEndian.Uint32(headbuf[6:])
	pkt.Command = binary.LittleEndian.Uint32(headbuf[10:])
	checksum := binary.LittleEndian.Uint32(headbuf[14:])

	if bodyLen > MaxAllowedV2CodecPayloadSize {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			pkt.Command, bodyLen, MaxAllowedV2CodecPayloadSize)
	}

	var nbytes = ServerCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(headbuf[:ServerCodecHeaderSize-4]); crc != checksum {
			return 0, errors.Errorf("message %d header checksum mismatch %x != %x",
				pkt.Command, checksum, crc)
		}
		return nbytes, nil
	}

	var payload = make([]byte, bodyLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, err
	}
	nbytes += bodyLen

	if pkt.Flag&fatchoy.PacketFlagEncrypted > 0 {
		if decryptor == nil {
			return 0, errors.Errorf("cannot decrypt message %d of size %d", pkt.Command, bodyLen)
		}
		payload = decryptor.Decrypt(payload)
	}

	hash := crc32.NewIEEE()
	hash.Write(headbuf[:ServerCodecHeaderSize-4])
	hash.Write(payload)
	if crc := hash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %d %d bytes checksum mismatch %x != %x",
			pkt.Command, bodyLen, checksum, crc)
	}
	pkt.Body = payload
	return nbytes, nil
}
