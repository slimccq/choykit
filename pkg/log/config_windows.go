// Copyright (C) 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build windows

package log

import (
	stdlog "log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
)

// Windows event log
type EventLogHook struct {
	ev   *eventlog.Log //
	name string        // name of logger
}

func NewEventLogHook(name string) Hooker {
	el, err := eventlog.Open(name)
	if err != nil {
		stdlog.Panicf("open event log %s: %v", name, err)
	}
	return &EventLogHook{
		name: name,
		ev:   el,
	}
}

func (h *EventLogHook) Name() string {
	return "eventlog"
}

func (h *EventLogHook) Fire(entry zapcore.Entry) error {
	switch entry.Level {
	case zapcore.PanicLevel:
		return h.ev.Error(3, entry.Message)
	case zapcore.FatalLevel:
		return h.ev.Error(3, entry.Message)
	case zapcore.ErrorLevel:
		return h.ev.Error(3, entry.Message)
	case zapcore.WarnLevel:
		return h.ev.Warning(2, entry.Message)
	case zapcore.InfoLevel:
		return h.ev.Info(1, entry.Message)
	case zapcore.DebugLevel:
		return h.ev.Info(1, entry.Message)
	default:
		return nil
	}
}

func platformSetup(name string) zap.Option {
	hook := NewEventLogHook(name)
	return zap.Hooks(hook.Fire)
}
