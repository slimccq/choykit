// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"hash/crc32"
	"io"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"github.com/pkg/errors"
)

const (
	ServerCodecVersion    = 2  // 协议版本
	ServerCodecHeaderSize = 18 // 消息头大小
)

var MaxAllowedServerCodecPayloadSize = 8 * 1024 * 1024 // 最大包体大小(8M)

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

// 把消息内容写入w，对消息的压缩和加密请在上层处理
func (c *serverProtocolCodec) Marshal(w io.Writer, pkt *fatchoy.Packet) error {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return err
	}
	if n := len(payload); n >= MaxAllowedServerCodecPayloadSize {
		pkt.Flag |= fatchoy.PacketFlagError
		var data [10]byte
		payload = fatchoy.EncodeNumber(protocol.ErrDataCodecFailure, data[:])
		log.Errorf("message %d too large payload %d/%d", pkt.Command, n, MaxAllowedServerCodecPayloadSize)
	}

	n := uint32(len(payload))
	hash := crc32.NewIEEE()
	var headbuf [ServerCodecHeaderSize]byte
	binary.LittleEndian.PutUint32(headbuf[0:], n)
	binary.LittleEndian.PutUint16(headbuf[4:], pkt.Flag)
	binary.LittleEndian.PutUint32(headbuf[6:], pkt.Seq)
	binary.LittleEndian.PutUint32(headbuf[10:], pkt.Command)
	hash.Write(headbuf[:ServerCodecHeaderSize-4])
	if n > 0 {
		hash.Write(payload)
	}
	binary.LittleEndian.PutUint32(headbuf[ServerCodecHeaderSize-4:], hash.Sum32())
	w.Write(headbuf[0:])
	if n > 0 {
		w.Write(payload)
	}
	return nil
}

// 从r中读取消息内容，只检查包体大小和校验码，压缩和解密请在之后处理
func (c *serverProtocolCodec) Unmarshal(r io.Reader, pkt *fatchoy.Packet) (int, error) {
	var headbuf [ServerCodecHeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return 0, err
	}

	bodyLen := int(binary.LittleEndian.Uint32(headbuf[0:]))
	pkt.Flag = binary.LittleEndian.Uint16(headbuf[4:])
	pkt.Seq = binary.LittleEndian.Uint32(headbuf[6:])
	pkt.Command = binary.LittleEndian.Uint32(headbuf[10:])
	checksum := binary.LittleEndian.Uint32(headbuf[14:])

	if bodyLen > MaxAllowedServerCodecPayloadSize {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			pkt.Command, bodyLen, MaxAllowedServerCodecPayloadSize)
	}

	var bytesRead = ServerCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(headbuf[:ServerCodecHeaderSize-4]); crc != checksum {
			return 0, errors.Errorf("message %d header checksum mismatch %x != %x",
				pkt.Command, checksum, crc)
		}
		return bytesRead, nil
	}

	var payload = make([]byte, bodyLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, err
	}
	bytesRead += bodyLen
	hash := crc32.NewIEEE()
	hash.Write(headbuf[:ServerCodecHeaderSize-4])
	hash.Write(payload)
	if crc := hash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %d %d bytes checksum mismatch %x != %x",
			pkt.Command, bodyLen, checksum, crc)
	}
	pkt.Body = payload
	return bytesRead, nil
}
