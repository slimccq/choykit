// Copyright © 2019-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Level int8
)

const (
	DebugLevel  = Level(zap.DebugLevel)
	InfoLevel   = Level(zap.InfoLevel)
	WarnLevel   = Level(zap.WarnLevel)
	ErrorLevel  = Level(zap.ErrorLevel)
	DPanicLevel = Level(zap.DPanicLevel)
	PanicLevel  = Level(zap.PanicLevel)
	FatalLevel  = Level(zap.FatalLevel)
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
