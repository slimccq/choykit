// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// 编码一个字符串、字节流、protobuf消息对象
// 编码后的字节用于传输，不能修改其内容
func EncodeValue(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		sh := (*reflect.StringHeader)(unsafe.Pointer(&v))
		bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}
		data := *(*[]byte)(unsafe.Pointer(&bh))
		return data, nil
	case proto.Message:
		data, err := proto.Marshal(v)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return data, nil
	default: // 使用GOB编码
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(value); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

// 编码数字, `data`需足够容量
func EncodeNumber(value interface{}, data []byte) []byte {
	buf := bytes.NewBuffer(data)
	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		panic(errors.Wrapf(err, "EncodeNumber: %v", value))
	}
	return data
}

// 解码Uint32
func DecodeU32(data []byte) (uint32, error) {
	var value uint32
	r := bytes.NewReader(data)
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

// 解码为string
func DecodeAsString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		bh := (*reflect.SliceHeader)(unsafe.Pointer(&v))
		sh := reflect.StringHeader{Data: bh.Data, Len: bh.Len}
		return *(*string)(unsafe.Pointer(&sh))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 解码为protobuf消息
func DecodeAsMsg(value interface{}, msg proto.Message) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if err := proto.Unmarshal(v, msg); err != nil {
			return errors.WithStack(err)
		}
	default:
		return errors.Errorf("invalid body type %T to decode", v)
	}
	return nil
}
