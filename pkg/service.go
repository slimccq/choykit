// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

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
