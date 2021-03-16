// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

const timestampLayout = "2006-01-02 15:04:05.999"

func Catch() {
	if v := recover(); v != nil {
		Backtrace(v, os.Stderr)
	}
}

func Backtrace(message interface{}, f *os.File) {
	if f == nil {
		f = os.Stderr
	}
	var buf bytes.Buffer
	var now = time.Now()
	fmt.Fprintf(&buf, "Traceback[%s] (most recent call last):\n", now.Format(timestampLayout))
	for i := 1; i < 50; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pc).Name()
		fmt.Fprintf(&buf, "% 2d. %s() %s:%d\n", i, funcName, file, line)
	}
	fmt.Fprintf(&buf, "%v\n", message)
	buf.WriteTo(f)
}
