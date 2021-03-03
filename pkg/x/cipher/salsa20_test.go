// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/rand"
	"testing"
)

func TestSalsa20_Decrypt(t *testing.T) {
	key := randBytes(32)
	nonce := uint64(1000)
	encryptor := NewSalsa20(nonce, key)
	decryptor := NewSalsa20(nonce, key)
	for i := 0; i < 100; i++ {
		payload := randBytes(100 + rand.Int()%1000)
		encrypted, _ := encryptor.Encrypt(cloneBytes(payload))
		decrypted, _ := decryptor.Decrypt(encrypted)
		if !bytes.Equal(payload, decrypted) {
			checksum1 := fmt.Sprintf("%x", md5.Sum(payload))
			checksum2 := fmt.Sprintf("%x", md5.Sum(decrypted))
			t.Fatalf("encryption mismatch %s != %s", checksum1, checksum2)
		}
	}
}
