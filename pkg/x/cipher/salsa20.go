// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"encoding/binary"
	"golang.org/x/crypto/salsa20"
)

// https://en.wikipedia.org/wiki/Salsa20
type salsa20Crypt struct {
	key   [32]byte
	nonce uint64
}

func NewSalsa20(key, nonce []byte) BlockCryptor {
	s := &salsa20Crypt{
		nonce: binary.BigEndian.Uint64(nonce[:8]),
	}
	copy(s.key[:], key)
	return s
}

func (s *salsa20Crypt) Encrypt(data []byte) []byte {
	var nonce [8]byte
	binary.BigEndian.PutUint64(nonce[:], s.nonce)
	s.nonce++
	salsa20.XORKeyStream(data, data, nonce[:], &s.key)
	return data
}

func (s *salsa20Crypt) Decrypt(data []byte) []byte {
	return s.Encrypt(data)
}
