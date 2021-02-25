// Copyright © 2017-present ichenq@outlook.com  All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package choykit

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

func Backtrace(message interface{}, fp *os.File) {
	if fp == nil {
		fp = os.Stderr
	}
	var buf bytes.Buffer
	var now = time.Now()
	fmt.Fprintf(&buf, "Traceback[%s] (most recent call last):\n", now.Format(timestampLayout))
	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i + 1)
		if !ok {
			break
		}
		fmt.Fprintf(&buf, "% 3d. %s() %s:%d\n", i, runtime.FuncForPC(pc).Name(), file, line)
	}
	fmt.Fprintf(&buf, "%v\n", message)
	buf.WriteTo(fp)
}
