// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

// 服务
type Service interface {
	ID() uint8
	Name() string
	NodeID() NodeID
	SetNodeID(NodeID)

	// 初始化、启动和关闭
	Init(*ServiceContext) error
	Startup() error
	Shutdown()

	// 服务上下文
	Context() *ServiceContext

	// 执行
	Execute(Runner) error
	Dispatch(*Packet) error
}
