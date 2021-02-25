// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	TimestampFormat = "2006-01-02 15:04:05.999"
)

var (
	logger *zap.Logger        // core logger
	sugar  *zap.SugaredLogger // sugared logger
	config zap.Config         // logger config
)

func IsTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

func setLogLevel(level string) {
	switch level {
	case "fatal":
		config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	case "panic":
		config.Level = zap.NewAtomicLevelAt(zap.PanicLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
}

func lazyInit() {
	if sugar == nil {
		Setup(false, false, "debug", "", "")
	}
}

func Setup(isProduction, enableSysLog bool, level, filepath, args string) {
	if isProduction {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	setLogLevel(level)
	if filepath != "" {
		config.OutputPaths = append(config.OutputPaths, filepath)
	}
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(TimestampFormat)
	if IsTerminal(os.Stdout) {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var options = []zap.Option{
		zap.AddCallerSkip(1),
	}
	if isProduction && enableSysLog {
		options = append(options, platformSetup(args))
	}
	logger, _ = config.Build(options...)
	sugar = logger.Sugar()
}
