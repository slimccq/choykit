// Copyright © 2020-present ichenq@outlook.com All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

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
