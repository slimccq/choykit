// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

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
	ClientCodecVersion    = 3         // 协议版本
	ClientCodecHeaderSize = 14        // 消息头大小
	MaxAllowedV1SendBytes = 60 * 1024 // 最大发送消息大小(60k)
	MaxAllowedV1RecvBytes = 16 * 1024 // 最大接收消息大小(16k)
)

//  header wire format
//       --------------------------------
// field | len | flag | seq | cmd | crc |
//       --------------------------------
// bytes |  2  |  2   |  2  |  4  |  4  |
//       --------------------------------

type clientCodec struct {
	enableSeqChk bool                        // 是否开启序列号检查
	lastSeq      uint16                      // 序列号
	encHash      hash.Hash32                 // 解码crc
	decHash      hash.Hash32                 // 编码crc
	headBuf      [ClientCodecHeaderSize]byte // 收包buffer
}

func NewClientCodec() *clientCodec {
	return &clientCodec{
		encHash: crc32.NewIEEE(),
		decHash: crc32.NewIEEE(),
	}
}

func (c *clientCodec) Version() uint8 {
	return ClientCodecVersion
}

func (c *clientCodec) Clone() choykit.Codec {
	return &clientCodec{
		encHash: crc32.NewIEEE(),
		decHash: crc32.NewIEEE(),
	}
}

func (c *clientCodec) SetSeqNo(seq uint16) {
	c.enableSeqChk = true
	c.lastSeq = seq
}

func (c *clientCodec) SetEncryptKey(key, iv []byte) {
	// TODO:
}

func (c *clientCodec) encodeHeader(pkt *choykit.Packet, length uint16, buffer *bytes.Buffer) {
	var tmpbuf [ClientCodecHeaderSize]byte
	binary.LittleEndian.PutUint16(tmpbuf[0:], length)
	binary.LittleEndian.PutUint16(tmpbuf[2:], pkt.Flags)
	binary.LittleEndian.PutUint16(tmpbuf[4:], pkt.Seq)
	binary.LittleEndian.PutUint32(tmpbuf[6:], pkt.Command)
	c.encHash.Write(tmpbuf[:ClientCodecHeaderSize-4])
	binary.LittleEndian.PutUint32(tmpbuf[10:], c.encHash.Sum32())
	buffer.Write(tmpbuf[0:])
}

// 编码消息
func (c *clientCodec) Encode(pkt *choykit.Packet, buf *bytes.Buffer) error {
	payload, err := pkt.Encode()
	if err != nil {
		return err
	}
	if (pkt.Flags & choykit.PacketFlagCompress) != 0 {
		if data, err := CompressBytes(ZLIB, DefaultCompression, payload); err != nil {
			log.Errorf("compress message %d: %v", pkt.Command, err)
			return err
		} else {
			payload = data
		}
	}
	if n := len(payload); n > MaxAllowedV1SendBytes {
		pkt.Flags |= choykit.PacketFlagError
		var data [10]byte
		payload = choykit.EncodeNumber(protocol.ErrDataCodecFailure, data[:])
		log.Errorf("message payload %v too large %d/%d", pkt.Command, n, MaxAllowedV1SendBytes)
	}

	c.encHash.Reset()
	n := len(payload)
	if n > 0 {
		c.encHash.Write(payload)
	}
	buf.Grow(ClientCodecHeaderSize + n)
	c.encodeHeader(pkt, uint16(n), buf)
	if n > 0 {
		buf.Write(payload)
	}
	return nil
}

func (c *clientCodec) decodeHeader(rd io.Reader, pkt *choykit.Packet, length *int, checksum *uint32) error {
	var buf = c.headBuf[:]
	if _, err := io.ReadFull(rd, buf); err != nil {
		return err
	}
	*length = int(binary.LittleEndian.Uint16(c.headBuf[0:]))
	pkt.Flags = binary.LittleEndian.Uint16(buf[2:])
	pkt.Seq = binary.LittleEndian.Uint16(buf[4:])
	pkt.Command = binary.LittleEndian.Uint32(buf[6:])
	*checksum = binary.LittleEndian.Uint32(c.headBuf[10:])
	return nil
}

// 解码消息
func (c *clientCodec) Decode(rd io.Reader, pkt *choykit.Packet) (int, error) {
	var bodyLen int
	var checksum uint32
	if err := c.decodeHeader(rd, pkt, &bodyLen, &checksum); err != nil {
		return 0, err
	}
	if bodyLen > MaxAllowedV1RecvBytes {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			bodyLen, bodyLen, MaxAllowedV1RecvBytes)
	}
	if c.enableSeqChk {
		if pkt.Seq != c.lastSeq {
			return 0, errors.Errorf("packet %v seq mismatch %d != %d", pkt.Command, pkt.Seq, c.lastSeq)
		}
		c.lastSeq++
	}
	var bytesRead = ClientCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(c.headBuf[:ClientCodecHeaderSize-4]); crc != checksum {
			return 0, errors.Errorf("message %v header checksum mismatch %x != %x",
				pkt.Command, checksum, crc)
		}
		return bytesRead, nil
	}
	var payload = make([]byte, bodyLen)
	if _, err := io.ReadFull(rd, payload); err != nil {
		return 0, err
	}
	bytesRead += bodyLen
	c.decHash.Reset()
	c.decHash.Write(payload)
	c.decHash.Write(c.headBuf[:ClientCodecHeaderSize-4])
	if crc := c.decHash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %v checksum mismatch: %x != %x", pkt.Command, checksum, crc)
	}

	if (pkt.Flags & choykit.PacketFlagCompress) != 0 {
		if data, err := UnCompressBytes(ZLIB, payload); err != nil {
			log.Errorf("uncompress message %d: %v", pkt.Command, err)
			return 0, err
		} else {
			payload = data
		}
	}
	if (pkt.Flags & choykit.PacketFlagError) != 0 {
		if ec, err := choykit.DecodeU32(payload); err != nil {
			return bytesRead, err
		} else {
			pkt.Body = ec
		}
	} else {
		pkt.Body = payload
	}
	return bytesRead, nil
}
