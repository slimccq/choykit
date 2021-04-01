// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package strutil

import (
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func ParseBool(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		return s[0] == '1'
	case 2:
		return s == "on" || s == "ON"
	case 3:
		return s == "yes" || s == "YES"
	case 4:
		return s == "true" || s == "TRUE"
	default:
		b, err := strconv.ParseBool(s)
		if err != nil {
			log.Panicf("ParseBool: cannot pasre %s to boolean: %v", s, err)
		}
		return b
	}
}

func ParseI8(s string) int8 {
	n := ParseI32(s)
	if n > math.MaxInt8 || n < math.MinInt8 {
		log.Panicf("ParseI8: value %s out of range", s)
	}
	return int8(n)
}

func ParseU8(s string) uint8 {
	n := ParseI32(s)
	if n > math.MaxUint8 || n < 0 {
		log.Panicf("ParseU8: value %s out of range", s)
	}
	return uint8(n)
}

func ParseI16(s string) int16 {
	n := ParseI32(s)
	if n > math.MaxInt16 || n < math.MinInt16 {
		log.Panicf("ParseI16: value %s out of range", s)
	}
	return int16(n)
}

func ParseU16(s string) uint16 {
	n := ParseI32(s)
	if n > math.MaxUint16 || n < 0 {
		log.Panicf("ParseU16: value %s out of range", s)
	}
	return uint16(n)
}

func ParseI32(s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		log.Panicf("ParseI32: cannot parse [%s] to int32: %v", s, err)
	}
	return int32(n)
}

func ParseU32(s string) uint32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		log.Panicf("ParseU32: cannot parse [%s] to uint32: %v", s, err)
	}
	return uint32(n)
}

func ParseI64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Panicf("ParseI64: cannot parse [%s] to int64: %v", s, err)
	}
	return n
}

func ParseU64(s string) uint64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Panicf("ParseU64: cannot parse [%s] to uint64: %v", s, err)
	}
	return n
}

func ParseF32(s string) float32 {
	f := ParseF64(s)
	if f > math.MaxFloat32 || f < math.SmallestNonzeroFloat32 {
		log.Panicf("ParseFloat32: value %s out of range", s)
	}
	return float32(f)
}

func ParseF64(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Panicf("ParseFloat64: cannot parse [%s] to double: %v", s, err)
	}
	return f
}

// 解析字符串为map，格式: "a='x,y',c=z" to {"a":"x,y", "c":"z"}
func ParseSepKeyValues(text string, sep1, sep2 string) map[string]string {
	const quote = "'"
	var result = make(map[string]string)
	for p := strings.Index(text, sep2); p > 0 && len(text) > 0; p = strings.Index(text, sep2) {
		key := strings.TrimSpace(text[:p])
		i := p + 1
		if i >= len(text) {
			result[key] = ""
			text = text[i:]
			continue
		}
		if text[i] == quote[0] {
			i++
			n := strings.Index(text[i:], quote)
			if n < 0 { // quote不配对
				break
			}
			value := text[i : i+n]
			result[key] = strings.TrimSpace(value)
			text = text[i+n+1:]
			n = strings.Index(text, sep1)
			if n < 0 {
				break // 最后一个kv
			}
			text = text[n+1:]
		} else {
			var value string
			n := strings.Index(text[i:], sep1)
			if n == 0 {
				result[key] = ""
			} else if n < 0 {
				value = text[i:]
				result[key] = strings.TrimSpace(value)
				break // 最后一个kv
			} else {
				value = text[i : i+n]
				result[key] = strings.TrimSpace(value)
			}
			text = text[i+n+1:]
		}
	}
	return result
}

// 解析字符串的值到value
func ParseStringToValue(s string, v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int8:
		v.SetInt(int64(ParseI8(s)))
	case reflect.Int16:
		v.SetInt(int64(ParseI16(s)))
	case reflect.Int32:
		v.SetInt(int64(ParseI32(s)))
	case reflect.Int:
		v.SetInt(ParseI64(s))
	case reflect.Int64:
		v.SetInt(ParseI64(s))

	case reflect.Uint8:
		v.SetUint(uint64(ParseU8(s)))
	case reflect.Uint16:
		v.SetUint(uint64(ParseU16(s)))
	case reflect.Uint32:
		v.SetUint(uint64(ParseU32(s)))
	case reflect.Uint:
		v.SetUint(ParseU64(s))
	case reflect.Uint64:
		v.SetUint(ParseU64(s))

	case reflect.Float32:
		v.SetFloat(float64(ParseF32(s)))
	case reflect.Float64:
		v.SetFloat(ParseF64(s))
	case reflect.Bool:
		v.SetBool(ParseBool(s))
	case reflect.String:
		v.SetString(s)

	default:
		return false
	}
	return true
}

// 解析字符串为数值、布尔
func ParseStringAs(typename, value string) interface{} {
	switch typename {
	case "int":
		return int(ParseI64(value))
	case "int8":
		return ParseI8(value)
	case "int16":
		return ParseI16(value)
	case "int32":
		return ParseI32(value)
	case "int64":
		return ParseI64(value)
	case "uint":
		return uint(ParseU64(value))
	case "uint8":
		return ParseU8(value)
	case "uint16":
		return ParseU16(value)
	case "uint32":
		return ParseU32(value)
	case "uint64":
		return ParseU64(value)
	case "float32":
		return ParseF32(value)
	case "float64":
		return ParseF64(value)
	case "bool":
		return ParseBool(value)
	}
	return value
}

// 将kv值设置到struct
func ParseKVToStruct(env map[string]string, ptr reflect.Value) {
	var sType = ptr.Type()
	for i := 0; i < ptr.NumField(); i++ {
		var name = sType.Field(i).Name
		if s, found := env[name]; found && s != "" {
			ParseStringToValue(s, ptr.Field(i))
		}
	}
}
