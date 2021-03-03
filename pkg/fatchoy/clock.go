// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"time"

	"devpkg.work/choykit/pkg/x/datetime"
)

var std *datetime.Clock // default clock

// 开启时钟
func StartClock() {
	std = datetime.NewClock(0)
	std.Go()
}

// 关闭时钟
func StopClock() {
	if std != nil {
		std.Stop()
		std = nil
	}
}

func WallClock() *datetime.Clock {
	return std
}

func Now() time.Time {
	return std.Now()
}

func DateTime() string {
	return std.DateTime()
}
