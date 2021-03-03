// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"encoding/binary"
	"golang.org/x/crypto/salsa20"
)

type Salsa20 struct {
	key   [32]byte
	nonce uint64
}

func NewSalsa20(nonce uint64, key []byte) *Salsa20 {
	s := &Salsa20{
		nonce: nonce,
	}
	copy(s.key[:], key)
	return s
}

func (s *Salsa20) Encrypt(data []byte) ([]byte, error) {
	var nonce [8]byte
	binary.BigEndian.PutUint64(nonce[:], s.nonce)
	s.nonce++
	salsa20.XORKeyStream(data, data, nonce[:], &s.key)
	return data, nil
}

func (s *Salsa20) Decrypt(data []byte) ([]byte, error) {
	return s.Encrypt(data)
}
