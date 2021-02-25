// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import (
	"testing"
	"time"
)

func TestStartClock(t *testing.T) {
	StartClock()
	defer StopClock()
	t.Logf("%s", DateTime())

	clock := WallClock()
	clock.Travel(-time.Hour)
	t.Logf("%s", DateTime())
}
