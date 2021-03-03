// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
)

// 压缩和加密
func EncodePacket(pkt *fatchoy.Packet, thresshold int, encrypt fatchoy.MessageEncryptor) error {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return err
	}
	if payload == nil {
		return nil
	}
	if n := len(payload); n > thresshold {
		if data, err := CompressBytes(payload); err != nil {
			log.Errorf("compress packet %d with %d bytes: %v", pkt.Command, n, err)
			return err
		} else {
			payload = data
			pkt.Flag |= fatchoy.PacketFlagCompressed
		}
	}
	if encrypt != nil {
		encrypted, er := encrypt.Encrypt(payload)
		if er != nil {
			log.Errorf("encrypt packet %d with %d bytes: %v", pkt.Command, len(payload), err)
			return er
		}
		payload = encrypted
		pkt.Flag |= fatchoy.PacketFlagEncrypted
	}
	pkt.Body = payload
	return nil
}

// 解密和解压缩
func DecodePacket(pkt *fatchoy.Packet, decrypt fatchoy.MessageEncryptor) error {
	payload, err := pkt.EncodeBody()
	if err != nil {
		return err
	}
	if payload == nil {
		return nil
	}
	if (pkt.Flag&fatchoy.PacketFlagEncrypted) != 0 && decrypt != nil {
		decrypted, err := decrypt.Decrypt(payload)
		if err != nil {
			return err
		}
		payload = decrypted
		pkt.Flag &= ^uint16(fatchoy.PacketFlagEncrypted)
	}

	if (pkt.Flag & fatchoy.PacketFlagCompressed) != 0 {
		if uncompressed, err := UncompressBytes(payload); err != nil {
			log.Errorf("uncompressed packet %d %d bytes: %v", pkt.Command, len(payload), err)
			return err
		} else {
			payload = uncompressed
			pkt.Flag &= ^uint16(fatchoy.PacketFlagCompressed)
		}
	}
	// 如果有FlagError，则body是32位数值错误码
	if (pkt.Flag & fatchoy.PacketFlagError) != 0 {
		if ec, err := fatchoy.DecodeU32(payload); err != nil {
			return err
		} else {
			pkt.Body = ec
		}
	} else {
		pkt.Body = payload
	}
	return nil
}
