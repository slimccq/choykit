// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"os"
	"testing"
)

func TestWriteFileLog(t *testing.T) {
	filename := "file.log"
	WriteFileLog(filename, "hello world")
	os.Remove(filename)
}

func TestServerErrorLog(t *testing.T) {
	// ServerErrorLog("server error log")
}

func TestFileLogger_Write(t *testing.T) {
	fw := NewFileWriter("prefix")
	defer os.Remove("prefix.log")
	fw.Write([]byte("Hello world"))
	defer fw.Sync()
}
