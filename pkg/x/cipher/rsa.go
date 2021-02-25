// Copyright © 2020-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// 解析公钥文件
func LoadRSAPublicKey(pemFile string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("incorrect public key file")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected key type %s", block.Type)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key.(*rsa.PublicKey), nil
}

// 解析私钥文件
func LoadRSAPrivateKey(pemFile string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("incorrect private key file")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected key type %s", block.Type)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// 最大加密内容大小
func MaxEncryptSize(pubkey *rsa.PublicKey) int {
	var k = pubkey.Size()
	var hash = sha256.New()
	return k - 2*hash.Size() - 2
}

func RSAEncrypt(pubkey *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubkey, data, nil)
}

func RSADecrypt(privkey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privkey, data, nil)
}
