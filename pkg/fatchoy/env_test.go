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

func TestParseNetInterface(t *testing.T) {
	tests := []struct {
		input            string
		expectedHostAddr string
		expectedBindAddr string
		expectedPort     int32
	}{
		{"", "", "", 0},
	}
	for _, tc := range tests {
		addr := ParseNetInterface(tc.input)
		if addr.AdvertiseAddr != tc.expectedHostAddr ||
			addr.BindAddr != tc.expectedBindAddr ||
			addr.Port != tc.expectedPort {
			t.Fatalf("unexpected addr: %v", addr)
		}
	}
}
