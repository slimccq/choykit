// Copyright (C) 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build linux

package log

import (
	stdlog "log"
	"log/syslog"

	"devpkg.work/choykit/pkg/x/strutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	writer  *syslog.Writer
	network string
	raddr   string
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
func NewSyslogHook(network, raddr string, priority syslog.Priority, tag string) Hooker {
	w, err := syslog.Dial(network, raddr, priority, tag)
	if err != nil {
		stdlog.Panicf("dial syslog: %v", err)
	}
	return &SyslogHook{
		writer:  w,
		network: network,
		raddr:   raddr,
	}
}

func (h *SyslogHook) Name() string {
	return "syslog"
}

func (h *SyslogHook) Fire(entry zapcore.Entry) error {
	var w = h.writer
	switch entry.Level {
	case zapcore.FatalLevel:
		return w.Emerg(entry.Message)
	case zapcore.PanicLevel:
		return w.Crit(entry.Message)
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		return w.Err(entry.Message)
	case zapcore.WarnLevel:
		return w.Warning(entry.Message)
	case zapcore.InfoLevel:
		return w.Info(entry.Message)
	case zapcore.DebugLevel:
		return w.Debug(entry.Message)
	default:
		return nil
	}
}

func toSysPriority(level zapcore.Level) syslog.Priority {
	switch level {
	case zapcore.FatalLevel:
		return syslog.LOG_EMERG
	case zapcore.PanicLevel:
		return syslog.LOG_CRIT
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		return syslog.LOG_ERR
	case zapcore.WarnLevel:
		return syslog.LOG_WARNING
	case zapcore.InfoLevel:
		return syslog.LOG_INFO
	case zapcore.DebugLevel:
		return syslog.LOG_DEBUG
	default:
		return syslog.LOG_INFO
	}
}

func platformSetup(args string) zap.Option {
	var network = "udp"
	var addr = "localhost:514"
	var priority = toSysPriority(config.Level.Level())
	var tag string
	kv := strutil.ParseSepKeyValues(args, ",", "=")
	if v, found := kv["network"]; found {
		network = v
	}
	if v, found := kv["addr"]; found {
		addr = v
	}
	if v, found := kv["priority"]; found {
		priority = syslog.Priority(strutil.ParseI32(v))
	}
	if v, found := kv["tag"]; found {
		tag = v
	}
	hook := NewSyslogHook(network, addr, priority, tag)
	return zap.Hooks(hook.Fire)
}
