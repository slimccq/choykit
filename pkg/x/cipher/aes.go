// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"log"
)

// https://en.wikipedia.org/wiki/Advanced_Encryption_Standard
type aesCrypt struct {
	encbuf [aes.BlockSize]byte
	decbuf [2 * aes.BlockSize]byte
	block  cipher.Block
	iv     []byte
}

func NewAES(key, iv []byte) *aesCrypt {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &aesCrypt{
		block: block,
		iv:    iv,
	}
}

func (c *aesCrypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *aesCrypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
