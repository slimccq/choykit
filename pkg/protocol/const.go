// Copyright © 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package protocol

const (
	InstanceAuto = 0          // 自动分配实例
	InstanceAll  = 0xFF       // 选取所有实例
	DistrictAll  = 0x0FFF     // 所有区服
	ReferAll     = 0xFFFFFFFF // 广播给所有session

	SERVICE_ALL = 0xFF // [FF]所有服务
)

// 与errno.proto一致
const (
	ErrDataCodecFailure      = 106 // 数据编码错误
	ErrRpcTimeout            = 108 // RPC超时
	ErrDuplicateRegistration = 201 // 服务重复注册
	ErrRegistrationDenied    = 202 // 服务注册被拒绝
)
