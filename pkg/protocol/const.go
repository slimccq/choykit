// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package protocol

const (
	InstanceAuto = 0          // 自动分配实例
	InstanceAll  = 0xFF       // 选取所有实例
	DistrictAll  = 0x0FFF     // 所有区服
	SessionAll   = 0xFFFFFFFF // 广播给所有session
	ServiceAll   = 0xFF       // 所有服务
)

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
	ErrDuplicateRegistration = 201 // 服务重复注册
	ErrRegistrationDenied    = 202 // 服务注册被拒绝
	ErrServerInternalError   = 203 // 服务器内部错误
	ErrServerMaintenance     = 204 // 服务器维护中
	ErrServiceNotAvailable   = 205 // 服务不可用
	ErrServiceBusy           = 206 // 服务正忙
	ErrDBException           = 207 // 数据库异常
	ErrSessionNotFound       = 208 // 未找到会话
)

// 服务类型
const (
	SERVICE_GATEWAY = 0x01	// 网关服务
)
