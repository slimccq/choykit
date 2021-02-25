// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"bytes"
	"testing"
)

// Run command below to generate test key files:
// 	openssl genrsa -out rsa_prikey.pem 1024
// 	openssl rsa -in rsa_prikey.pem -pubout -out rsa_pubkey.pem

func TestRSADecrypt(t *testing.T) {
	prikey, err := LoadRSAPrivateKey("rsa_prikey.pem")
	if err != nil {
		t.Fatalf("load private key: %v", err)
	}
	pubkey, err := LoadRSAPublicKey("rsa_pubkey.pem")
	if err != nil {
		t.Fatalf("load public key: %v", err)
	}
	var maxSize = MaxEncryptSize(pubkey)
	var data = []byte("a quick brown fox jumps over the lazy dog")
	encrypted, err := RSAEncrypt(pubkey, data)
	if err != nil {
		t.Fatalf("RSAEncrypt: %v, %d/%d", err, len(data), maxSize)
	}
	decrypted, err := RSADecrypt(prikey, encrypted)
	if err != nil {
		t.Fatalf("RSADecrypt: %v", err)
	}
	if !bytes.Equal(data, decrypted) {
		t.Fatalf("data not equal after encryption/decription")
	}
	t.Logf("RSA encryption OK")
}
