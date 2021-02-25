// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

// Runner是一个可执行对象接口
type Runner interface {
	Run() error
}

type Runnable struct {
	F func() error
}

func (r *Runnable) Run() error {
	return r.F()
}

func NewRunner(f func() error) Runner {
	return &Runnable{
		F: f,
	}
}

type CapturedRunnable struct {
	F func() error
}

func (r *CapturedRunnable) Run() error {
	defer Catch()
	return r.F()
}

func NewCapturedRunnable(f func() error) Runner {
	return &CapturedRunnable{
		F: f,
	}
}
