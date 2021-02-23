// Copyright © 2021 ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package reflectutil

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
)

// 获取struct内所有field的值
func GetStructAllFieldValues(ptr interface{}) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		var field = value.Field(i)
		result = append(result, field.Interface())
	}
	return result
}

func GetStructFieldValues(ptr interface{}, except string) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, st.NumField())
	for i := 0; i < st.NumField(); i++ {
		var field = value.Field(i)
		result = append(result, field.Interface())
	}
	return result
}

// 获取struct内指定field的值
func GetStructFieldValuesBy(ptr interface{}, names []string) []interface{} {
	var value = reflect.ValueOf(ptr).Elem()
	var st = value.Type()
	var result = make([]interface{}, 0, len(names))
	for _, fname := range names {
		_, ok := st.FieldByName(fname)
		if !ok {
			log.Panicf("field %s.%s not found", st.Name(), fname)
			return nil
		}
		field := value.FieldByName(fname)
		result = append(result, field.Interface())
	}
	return result
}

// 调用struct内指定名称的函数
func CallObjectMethod(ptr interface{}, method string, args ...interface{}) []interface{} {
	value := reflect.ValueOf(ptr)
	fn := value.MethodByName(method)
	var input []reflect.Value
	for _, arg := range args {
		input = append(input, reflect.ValueOf(arg))
	}
	output := fn.Call(input)
	var result []interface{}
	for _, out := range output {
		result = append(result, out.Interface())
	}
	return result
}

func DeepCopy(src, dst interface{}) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return err
	}
	dec := gob.NewDecoder(buf)
	return dec.Decode(dst)
}
