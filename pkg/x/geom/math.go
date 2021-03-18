// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

func IntMax(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func IntMin(x, y int) int {
	if y < x {
		return y
	}
	return x
}