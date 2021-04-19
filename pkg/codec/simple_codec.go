package codec

import (
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/protocol"
	"encoding/binary"
	"github.com/pkg/errors"
	"hash/crc32"
	"io"
)

const (
	SimpleCodecHeaderSize = 12 // 消息头大小
)

var MaxAllowedSimpleCodecPayloadSize = 8 * 1024 * 1024 // 最大包体大小(8M)

//  header wire format
//       --------------------------
// field | len | flag | cmd | crc |
//       --------------------------
// bytes |  3  |  1   |  4  |  4  |
//       --------------------------

// 编码器
type simpleProtocolCodec struct {
}

// 把消息内容写入w，对消息的压缩和加密请在上层处理
func (c *simpleProtocolCodec) Marshal(w io.Writer, pkt *fatchoy.Packet) error {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return err
	}
	if n := len(payload); n >= MaxAllowedSimpleCodecPayloadSize {
		pkt.Flag |= fatchoy.PacketFlagError
		var data [10]byte
		payload = fatchoy.EncodeNumber(protocol.ErrDataCodecFailure, data[:])
		log.Errorf("message %d too large payload %d/%d", pkt.Command, n, MaxAllowedSimpleCodecPayloadSize)
	}

	n := uint32(len(payload))
	hash := crc32.NewIEEE()
	var headbuf [SimpleCodecHeaderSize]byte
	binary.LittleEndian.PutUint32(headbuf[0:], n)
	headbuf[3] = uint8(pkt.Flag)
	binary.LittleEndian.PutUint32(headbuf[4:], pkt.Command)
	hash.Write(headbuf[:SimpleCodecHeaderSize-4])
	if n > 0 {
		hash.Write(payload)
	}
	binary.LittleEndian.PutUint32(headbuf[SimpleCodecHeaderSize-4:], hash.Sum32())
	w.Write(headbuf[0:])
	if n > 0 {
		w.Write(payload)
	}
	return nil
}

// 从r中读取消息内容，只检查包体大小和校验码，压缩和解密请在之后处理
func (c *simpleProtocolCodec) Unmarshal(r io.Reader, pkt *fatchoy.Packet) (int, error) {
	var headbuf [SimpleCodecHeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return 0, err
	}

	bodyLen := int(binary.LittleEndian.Uint32(headbuf[0:]))
	pkt.Flag = uint16(bodyLen >> 24)
	bodyLen &= 0x00FFFFFF

	pkt.Command = binary.LittleEndian.Uint32(headbuf[4:])
	checksum := binary.LittleEndian.Uint32(headbuf[8:])
	if bodyLen > MaxAllowedSimpleCodecPayloadSize {
		return 0, errors.Errorf("packet %d payload size overflow %d/%d",
			pkt.Command, bodyLen, MaxAllowedSimpleCodecPayloadSize)
	}

	var bytesRead = SimpleCodecHeaderSize
	if bodyLen == 0 {
		if crc := crc32.ChecksumIEEE(headbuf[:SimpleCodecHeaderSize-4]); crc != checksum {
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
	hash.Write(headbuf[:SimpleCodecHeaderSize-4])
	hash.Write(payload)
	if crc := hash.Sum32(); checksum != crc {
		return 0, errors.Errorf("message %d %d bytes checksum mismatch %x != %x",
			pkt.Command, bodyLen, checksum, crc)
	}
	pkt.Body = payload
	return bytesRead, nil
}
