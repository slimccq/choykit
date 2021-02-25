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

// 与errno.proto一致
const (
	ErrDataCodecFailure      = 106 // 数据编码错误
	ErrRpcTimeout            = 108 // RPC超时
	ErrDuplicateRegistration = 201 // 服务重复注册
	ErrRegistrationDenied    = 202 // 服务注册被拒绝
)
