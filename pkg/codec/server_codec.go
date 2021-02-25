// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"bytes"
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"

	"devpkg.work/choykit/pkg"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"github.com/pkg/errors"
)

const (
	ServerCodecVersion      = 1               // 协议版本
	ServerCodecHeaderSize   = 24              // 消息头大小
	MaxAllowedV2PayloadSize = 8 * 1024 * 1024 // 最大包体大小(8M)
)

// wire format of codec header
//       -----------------------------------------------
// field | len | flag | seq | cmd | node | refer | crc |
//       -----------------------------------------------
// bytes |  4  |  2   |  2  |  4  |   4  |   4   |  4  |
//       -----------------------------------------------

type serverHeader struct {
	Len      uint32 // 消息体大小
	Flag     uint16 //
	Seq      uint16 // 序列号
	MsgID    uint32 // 消息ID
	Node     uint32 // 目标节点号
	Refer    uint32 // session引用
	Checksum uint32 // CRC
}

// 编码器
type serverCodec struct {
	enableSeqChk bool                        //
	lastSeq      uint16                      // 序列号校验
	encHash      hash.Hash32                 // 解码crc
	decHash      hash.Hash32                 // 编码crc
	headBuf      [ServerCodecHeaderSize]byte // 收包buffer
}

func NewServerCodec() *serverCodec {
	return &serverCodec{
		encHash: crc32.NewIEEE(),
		decHash: crc32.NewIEEE(),
	}
}

func (c *serverCodec) Version() uint8 {
	return ServerCodecVersion
}

func (c *serverCodec) Clone() choykit.Codec {
	return &serverCodec{
		encHash: crc32.NewIEEE(),
		decHash: crc32.NewIEEE(),
	}
}

func (c *serverCodec) SetSeqNo(seq uint16) {
	c.enableSeqChk = true
	c.lastSeq = seq
}

func (c *serverCodec) SetEncryptKey(key, iv []byte) {
	// TODO:
}

func (c *serverCodec) encodeHeader(pkt *choykit.Packet, length uint32, buffer *bytes.Buffer) {
	var tmpbuf [ServerCodecHeaderSize]byte
	binary.LittleEndian.PutUint32(tmpbuf[0:], length)
	binary.LittleEndian.PutUint16(tmpbuf[4:], pkt.Flags)
	binary.LittleEndian.PutUint16(tmpbuf[6:], pkt.Seq)
	binary.LittleEndian.PutUint32(tmpbuf[8:], pkt.Command)
	binary.LittleEndian.PutUint32(tmpbuf[12:], uint32(pkt.Node))
	binary.LittleEndian.PutUint32(tmpbuf[16:], pkt.Referer)
	c.encHash.Write(tmpbuf[:ServerCodecHeaderSize-4])
	binary.LittleEndian.PutUint32(tmpbuf[20:], c.encHash.Sum32())
	buffer.Write(tmpbuf[0:])
}

// 编码
func (c *serverCodec) Encode(pkt *choykit.Packet, buf *bytes.Buffer) error {
	payload, err := pkt.Encode()
	if err != nil {
		return err
	}
	if n := len(payload); n >= MaxAllowedV2PayloadSize {
		pkt.Flags |= choykit.PacketFlagError
		var data [10]byte
		payload = choykit.EncodeNumber(protocol.ErrDataCodecFailure, data[:])
		log.Errorf("message %d too large payload %d/%d", pkt.Command, n, MaxAllowedV2PayloadSize)
	}

	if (pkt.Flags & choykit.PacketFlagCompress) != 0 {
		if data, err := CompressBytes(ZLIB, DefaultCompression, payload); err != nil {
			log.Errorf("compress message %d: %v", pkt.Command, err)
			return err
		} else {
			payload = data
		}
	}
	c.encHash.Reset()
	n := len(payload)
	if n > 0 {
		c.encHash.Write(payload)
	}
	c.encodeHeader(pkt, uint32(n), buf)
	if n > 0 {
		buf.Write(payload)
	}
	return nil
}

func (c *serverCodec) decodeHeader(rd io.Reader, pkt *choykit.Packet, length *int, checksum *uint32) error {
	var buf = c.headBuf[0:]
	if _, err := io.ReadFull(rd, buf); err != nil {
		return err
	}
	*length = int(binary.LittleEndian.Uint32(buf[0:]))
	pkt.Flags = binary.LittleEndian.Uint16(buf[4:])
	pkt.Seq = binary.LittleEndian.Uint16(buf[6:])
	pkt.Command = binary.LittleEndian.Uint32(buf[8:])
	pkt.Node = choykit.NodeID(binary.LittleEndian.Uint32(buf[12:]))
	pkt.Referer = binary.LittleEndian.Uint32(buf[16:])
	*checksum = binary.LittleEndian.Uint32(buf[20:])
	return nil
}

// 解包
func (c *serverCodec) Decode(rd io.Reader, pkt *choykit.Packet) (int, error) {
	var bodyLen int
	var checksum uint32
	if err := c.decodeHeader(rd, pkt, &bodyLen, &checksum); err != nil {
		return 0, err
	}
	if bodyLen > MaxAllowedV2PayloadSize {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			pkt.Command, bodyLen, MaxAllowedV2PayloadSize)
	}
	if c.enableSeqChk {
		if pkt.Seq != c.lastSeq {
			return 0, errors.Errorf("packet %d seq mismatch %d/%d", pkt.Command, pkt.Seq, c.lastSeq)
		}
		c.lastSeq++
	}
	var bytesRead = ServerCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(c.headBuf[:ServerCodecHeaderSize-4]); crc != checksum {
			return 0, errors.Errorf("message %d header checksum mismatch %x != %x",
				pkt.Command, checksum, crc)
		}
		return bytesRead, nil
	}

	var bodyData = make([]byte, bodyLen)
	if _, err := io.ReadFull(rd, bodyData); err != nil {
		return 0, err
	}
	bytesRead += bodyLen
	c.decHash.Reset()
	c.decHash.Write(bodyData)
	c.decHash.Write(c.headBuf[:ServerCodecHeaderSize-4])
	if crc := c.decHash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %d %d bytes checksum mismatch %x != %x",
			pkt.Command, bodyLen, checksum, crc)
	}
	if (pkt.Flags & choykit.PacketFlagCompress) != 0 {
		if data, err := UnCompressBytes(ZLIB, bodyData); err != nil {
			log.Errorf("message %d uncompress: %v", pkt.Command, err)
			return 0, err
		} else {
			bodyData = data
		}
	}
	if (pkt.Flags & choykit.PacketFlagError) != 0 {
		if n, err := binary.ReadVarint(bytes.NewReader(bodyData)); err != nil {
			return bytesRead, err
		} else {
			pkt.Body = uint32(n)
		}
	} else {
		pkt.Body = bodyData
	}
	return bytesRead, nil
}
