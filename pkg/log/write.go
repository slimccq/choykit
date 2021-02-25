// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"devpkg.work/choykit/pkg/x/fsutil"
	"golang.org/x/crypto/ssh/terminal"
)

func WriteFileLog(filename, format string, a ...interface{}) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, format, a...)
	return err
}

func AppFileErrorLog(format string, a ...interface{}) error {
	_, appname := filepath.Split(os.Args[0])
	var filename = fmt.Sprintf("%s_error.log", appname)
	return WriteFileLog(filename, format, a...)
}

// 用于在初始化服务时的错误日志
func ServerErrorLog(format string, a ...interface{}) {
	var isTerminal = terminal.IsTerminal(int(os.Stderr.Fd()))
	if isTerminal {
		fmt.Fprintf(os.Stderr, format, a...)
	} else {
		AppFileErrorLog(format, a...)
	}
}

// 记录日志到文件
type FileLogger struct {
	filename string
	writer   *fsutil.FileWriter
}

func NewFileWriter(prefix string) *FileLogger {
	var filename = fmt.Sprintf("%s.log", prefix)
	var w = fsutil.NewFileWriter(filename, 0)
	if err := w.Init(); err != nil {
		log.Panicf("FileWriter.Init: %v", err)
	}
	return &FileLogger{
		filename: filename,
		writer:   w,
	}
}

func (w *FileLogger) Write(data []byte) error {
	_, err := w.writer.Write(data)
	return err
}

func (w *FileLogger) Sync() error {
	return w.writer.Flush()
}
