// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import (
	"time"

	"devpkg.work/choykit/pkg/x/datetime"
)

var std *datetime.Clock // default clock

func StartClock() {
	std = datetime.NewClock(0)
	std.Go()
}

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
