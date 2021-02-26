// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"fmt"
	"testing"
)

func TestEnvExample(t *testing.T) {
	env := LoadEnviron()
	fmt.Printf("%v\n", env)
}
