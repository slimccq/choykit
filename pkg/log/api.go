// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"go.uber.org/zap/zapcore"
)

type Hooker interface {
	Name() string
	Fire(entry zapcore.Entry) error
}

func Shutdown() {
	logger.Sync()
}

func Debugf(format string, args ...interface{}) {
	lazyInit()
	sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	lazyInit()
	sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	lazyInit()
	sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	lazyInit()
	sugar.Errorf(format, args...)
}

func DPanicf(format string, args ...interface{}) {
	lazyInit()
	sugar.DPanicf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	lazyInit()
	sugar.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	lazyInit()
	sugar.Fatalf(format, args...)
}
