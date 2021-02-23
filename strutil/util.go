// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package strutil

import (
	"fmt"
	"math/rand"
	"reflect"
	"unicode"
	"unsafe"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-=~!@#$%^&*()_+[]\\;',./{}|:<>?"

// 对[]byte的修改会影响到返回的string
func FastBytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{Data: bh.Data, Len: bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// 修改返回的[]byte会引起panic
func FastStringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: 0}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func RandString(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx := rand.Int() % len(alphabet)
		result[i] = alphabet[idx]
	}
	return string(result)
}

func RandBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		ch := uint8(rand.Int31() % 0xFF)
		result[i] = ch
	}
	return result
}

// 在array中查找string
func FindStringInArray(a []string, x string) int {
	for i, v := range a {
		if v == x {
			return i
		}
	}
	return -1
}

// 查找第一个数字的位置
func FindFirstDigit(s string) int {
	for i, r := range s {
		if unicode.IsDigit(r) {
			return i
		}
	}
	return -1
}

// 查找第一个非数字的位置
func FindFirstNonDigit(s string) int {
	for i, r := range s {
		if !unicode.IsDigit(r) {
			return i
		}
	}
	return -1
}

// reverses the string
func Reverse(str string) string {
	runes := []rune(str)
	l := len(runes)
	for i := 0; i < l/2; i++ {
		runes[i], runes[l-i-1] = runes[l-i-1], runes[i]
	}
	return string(runes)
}

// pretty bytes to string
func PrettyBytes(n int64) string {
	if n < (1 << 10) {
		return fmt.Sprintf("%dB", n)
	} else if n < (1 << 20) {
		return fmt.Sprintf("%.2fKB", float64(n)/(1<<10))
	} else if n < (1 << 30) {
		return fmt.Sprintf("%.2fMB", float64(n)/(1<<20))
	} else if n < (1 << 40) {
		return fmt.Sprintf("%.2fGB", float64(n)/(1<<30))
	} else {
		return fmt.Sprintf("%.2fTB", float64(n)/(1<<40))
	}
}

// s1和s2的最长共同前缀
func LongestCommonPrefix(s1, s2 string) string {
	if s1 == "" || s2 == "" {
		return ""
	}
	i := 0
	for i < len(s1) && i < len(s2) {
		if s1[i] == s2[i] {
			i++
			continue
		}
		break
	}
	return s1[:i]
}
