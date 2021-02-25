// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

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
