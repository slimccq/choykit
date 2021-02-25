// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import "testing"

func TestServiceContext(t *testing.T) {
	env := LoadEnviron()
	opt := NewOptions()
	ctx := NewServiceContext(opt, env)
	ctx.Go()
}
