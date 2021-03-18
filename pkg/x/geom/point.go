// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

import (
	"math"
)

// 点
type Point struct {
	X, Y int
}

func NewPoint(x, y int) *Point {
	return &Point{x, y}
}

// 两点之间的距离
func (a *Point) Distance(b *Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(float64(dx), float64(dy))
}

func (a *Point) Equal(b *Point) bool {
	return a.X == b.X && a.Y == b.Y
}
