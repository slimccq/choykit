// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package protocol

import (
	"fmt"
	"hash/fnv"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	idNames     = make(map[uint32]string)
	nameIds     = make(map[string]uint32)
	msgRegistry = make(map[string]reflect.Type)
)

// 消息协议规则:
//  1, 请求消息以Req结尾
//  2, 响应消息以Ack结尾
//  3, 通知消息以Ntf结尾
func hasValidSuffix(name string) bool {
	nameSuffix := []string{"Req", "Ack", "Ntf"}
	for _, suffix := range nameSuffix {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func messageNameHash(name string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(name))
	return h.Sum32()
}

func isNil(c interface{}) bool {
	if c == nil {
		return true
	}
	return reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()
}

// 根据消息名称的hash注册
func registerByNameHash(fileDescriptor protoreflect.FileDescriptor) bool {
	fmt.Printf("register %s\n", fileDescriptor.Path())
	msgDescriptors := fileDescriptor.Messages()
	for i := 0; i < msgDescriptors.Len(); i++ {
		descriptor := msgDescriptors.Get(i)
		fullname := descriptor.FullName()
		if !hasValidSuffix(string(fullname)) {
			continue
		}
		name := string(fullname)
		hash := messageNameHash(name)
		rtype := proto.MessageType(string(fullname))
		if s, found := idNames[hash]; found {
			log.Panicf("duplicate message hash %s %s, %d", s, name, hash)
		}
		msgRegistry[name] = rtype.Elem()
		nameIds[name] = hash
		idNames[hash] = name
	}
	return true
}

// 从message的option里获取消息ID
func getMsgIdByExtension(descriptor protoreflect.MessageDescriptor, ext *proto.ExtensionDesc) uint32 {
	ovi := descriptor.Options()
	if isNil(ovi) {
		return 0
	}
	omi := ovi.ProtoReflect()
	var msgId uint32
	fullname := ext.TypeDescriptor().FullName()
	omi.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}
		if fd.FullName() == fullname {
			ivs := v.String()
			n, _ := strconv.ParseInt(ivs, 10, 64)
			msgId = uint32(n)
			return false
		}
		return true
	})
	return msgId
}

// 根据消息option指定的ID注册
func registerByExtension(fileDescriptor protoreflect.FileDescriptor, ext *proto.ExtensionDesc) bool {
	fmt.Printf("register %s\n", fileDescriptor.Path())
	msgDescriptors := fileDescriptor.Messages()
	for i := 0; i < msgDescriptors.Len(); i++ {
		descriptor := msgDescriptors.Get(i)
		fullname := descriptor.FullName()
		if !hasValidSuffix(string(fullname)) {
			continue
		}
		name := string(fullname)
		msgid := getMsgIdByExtension(descriptor, ext)
		if msgid == 0 {
			continue
		}
		rtype := proto.MessageType(string(fullname))
		if s, found := idNames[msgid]; found {
			log.Panicf("duplicate message hash %s %s, %d", s, name, msgid)
		}
		msgRegistry[name] = rtype.Elem()
		nameIds[name] = msgid
		idNames[msgid] = name
	}
	return true
}

func RegisterV1() {
	protoregistry.GlobalFiles.RangeFiles(registerByNameHash)
	fmt.Printf("%d messages registered\n", len(msgRegistry))
}

func RegisterV2(ext *proto.ExtensionDesc) {
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		return registerByExtension(fd, ext)
	})
	fmt.Printf("%d messages registered\n", len(msgRegistry))
}

// 根据消息ID创建message
func CreateMessageBy(msgId uint32) proto.Message {
	if name, found := idNames[msgId]; found {
		if rtype, ok := msgRegistry[name]; ok {
			msg := reflect.New(rtype).Interface()
			return msg.(proto.Message)
		}
	}
	return nil
}

// 根据message获取消息ID
func GetMessageIDOf(msg proto.Message) uint32 {
	rtype := reflect.TypeOf(msg)
	fullname := rtype.String()
	if fullname == "" {
		return 0
	}
	for fullname[0] == '*' {
		fullname = fullname[1:]
	}
	return nameIds[fullname]
}
