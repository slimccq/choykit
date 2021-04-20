// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package protocol


// 错误码，定义与errno.proto一致
const (
	ErrBadRequest            = 101 // 错误的请求
	ErrInvalidArgument       = 102 // 参数错误
	ErrOperationNotSupported = 103 // 不支持当前操作
	ErrOperationTooOften     = 104 // 操作过于频繁
	ErrRequestTimeout        = 105 // 请求超时
	ErrDataCodecFailure      = 106 // 数据编码错误
	ErrProtocolIncompatible  = 107 // 协议不兼容
	ErrRpcTimeout            = 108 // RPC超时
)
