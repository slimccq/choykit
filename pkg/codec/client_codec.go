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
	ClientCodecVersion    = 2  // 协议版本
	ClientCodecHeaderSize = 16 // 消息头大小
)

var (
	MaxAllowedV1SendBytes = 60 * 1024 // 最大发送消息大小(60k)
	MaxAllowedV1RecvBytes = 8 * 1024  // 最大接收消息大小(8k)
)

//  header wire format
//       --------------------------------
// field | len | flag | seq | cmd | crc |
//       --------------------------------
// bytes |  2  |  2   |  4  |  4  |  4  |
//       --------------------------------

type clientProtocolCodec struct {
}

var ClientProtocolCodec = NewClientProtocolCodec()

func NewClientProtocolCodec() fatchoy.ProtocolCodec {
	return &clientProtocolCodec{}
}

// 把消息内容写入w，并按需对body加密，消息压缩请在上层处理
func (c *clientProtocolCodec) Marshal(w io.Writer, encryptor cipher.BlockCryptor, pkt *fatchoy.Packet) (int, error) {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return 0, err
	}
	if len(payload) > 0 && encryptor != nil {
		payload = encryptor.Encrypt(payload)
		pkt.Flag |= fatchoy.PacketFlagEncrypted
	}
	if n := len(payload); n > MaxAllowedV1SendBytes {
		return 0, errors.Errorf("message payload %v too large %d/%d", pkt.Command, n, MaxAllowedV1SendBytes)
	}

	hash := crc32.NewIEEE()
	n := len(payload)
	var headbuf [ClientCodecHeaderSize]byte
	binary.LittleEndian.PutUint16(headbuf[0:], uint16(n))
	binary.LittleEndian.PutUint16(headbuf[2:], pkt.Flag)
	binary.LittleEndian.PutUint32(headbuf[4:], pkt.Seq)
	binary.LittleEndian.PutUint32(headbuf[8:], pkt.Command)
	hash.Write(headbuf[:ClientCodecHeaderSize-4])
	if n > 0 {
		hash.Write(payload)
	}
	binary.LittleEndian.PutUint32(headbuf[ClientCodecHeaderSize-4:], hash.Sum32())
	nbytes, err := w.Write(headbuf[0:])
	if err == nil && n > 0 {
		n, err = w.Write(payload)
		nbytes += n
	}
	return nbytes, err
}

// 从r中读取消息内容，检查包体大小和校验码，和解密，解压缩请在之后处理
func (c *clientProtocolCodec) Unmarshal(r io.Reader, decryptor cipher.BlockCryptor, pkt *fatchoy.Packet) (int, error) {
	var headbuf [ClientCodecHeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return 0, err
	}
	bodyLen := int(binary.LittleEndian.Uint16(headbuf[0:]))
	pkt.Flag = binary.LittleEndian.Uint16(headbuf[2:])
	pkt.Seq = binary.LittleEndian.Uint32(headbuf[4:])
	pkt.Command = binary.LittleEndian.Uint32(headbuf[8:])
	checksum := binary.LittleEndian.Uint32(headbuf[12:])

	if bodyLen > MaxAllowedV1RecvBytes {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			pkt.Command, bodyLen, MaxAllowedV1RecvBytes)
	}
	var nbytes = ClientCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(headbuf[:ClientCodecHeaderSize-4]); crc != checksum {
			return 0, errors.Errorf("message %v header checksum mismatch %x != %x",
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
	hash.Write(headbuf[:ClientCodecHeaderSize-4])
	hash.Write(payload)
	if crc := hash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %v checksum mismatch: %x != %x", pkt.Command, checksum, crc)
	}
	pkt.Body = payload
	return nbytes, nil
}
