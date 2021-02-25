// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package protocol

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// 自定义协议消息message定义规则:
//  1, 请求消息以Req结尾
//  2, 响应消息以Ack结尾
//  3, 通知消息以Ntf结尾
var (
	pbMsgSuffix = []string{"Req", "Ack", "Ntf"}
	registry    = make(map[int32]reflect.Type)
	revRegistry = make(map[reflect.Type]int32)
)

// 根据反射拿到每一个Message的消息ID
func RetrieveFileMessages(ext *proto.ExtensionDesc, filename string) error {
	gzippedData := proto.FileDescriptor(filename)
	if gzippedData == nil {
		return fmt.Errorf("file %s not found in proto registry", filename)
	}
	rd, err := gzip.NewReader(bytes.NewReader(gzippedData))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rd); err != nil {
		return err
	}
	if err := rd.Close(); err != nil {
		return err
	}
	var fileDesc descriptor.FileDescriptorProto
	if err := proto.Unmarshal(buf.Bytes(), &fileDesc); err != nil {
		return err
	}
	for _, md := range fileDesc.MessageType {
		if !hasValidSuffix(*md.Name) {
			continue
		}
		var name = fmt.Sprintf("%s.%s", *fileDesc.Package, *md.Name)
		refType := proto.MessageType(name)
		v, err := proto.GetExtension(md.Options, ext)
		if err != nil {
			log.Fatalf("message ID extension not found for %s: %s", name, err)
			continue
		}
		msgid := *v.(*int32)
		if v, found := registry[msgid]; found {
			log.Fatalf("duplicate message ID found %v for %s and %s", msgid, v.Name(), refType.Name())
			continue
		}
		registry[msgid] = refType
		revRegistry[refType] = msgid
	}
	return nil
}

func hasValidSuffix(name string) bool {
	for _, suffix := range pbMsgSuffix {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// 根据消息ID创建message
func CreateMessageBy(msgId int32) proto.Message {
	if refType, ok := registry[msgId]; ok {
		ptr := reflect.Zero(refType)
		msg := ptr.Interface().(proto.Message)
		return msg
	}
	return nil
}

// 根据message获取消息ID
func GetMessageIDOf(msg proto.Message) int32 {
	refType := reflect.TypeOf(msg)
	return revRegistry[refType]
}

// 注册消息反射
func InitMsgRegistry(ext *proto.ExtensionDesc, pbFiles ...string) {
	for _, filename := range pbFiles {
		if err := RetrieveFileMessages(ext, filename); err != nil {
			log.Fatalf("RetrieveFileMessages %s: %v", filename, err)
		}
	}
	log.Printf("%d messages registered\n", len(registry))
}
