// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
)

// Two dimensional point
type Point2 struct {
	X, Y int
}

var EmptyPoint Point2

// Allocates and returns a new 2D point
func NewPoint2(x, y int) *Point2 {
	return &Point2{x, y}
}

// Calculates the distance between x and y
func (a *Point2) Dist(b *Point2) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(float64(dx), float64(dy))
}

// Returns whether two points are equal
func (a *Point2) Equal(b *Point2) bool {
	return a.X == b.X && a.Y == b.Y
}
