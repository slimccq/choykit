// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package choykit

import "testing"

func TestRouter(t *testing.T) {
	endpoints := NewEndpointMap()
	router := NewRouter(1234)
	policy := NewBasicRoutePolicy(endpoints)
	router.AddPolicy(policy)
}
