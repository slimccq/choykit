// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
)

// Two dimensional vector.
type Vec2 struct {
	X, Y int
}

// Allocates and returns a new 2D vector.
func NewVec2ByPoint(a, b Point2) *Vec2 {
	return &Vec2{
		X: b.X - a.X,
		Y: b.Y - a.Y,
	}
}

// Returns the length of x.
func (a *Vec2) Norm() float64 {
	return math.Hypot(float64(a.X), float64(a.Y))
}

// Normalize returns a unit point in the same direction as p.
func (a *Vec2) Normalize() Vec2 {
	if a.X == 0 && a.Y == 0 {
		return *a
	}
	return a.Mul(1 / a.Norm())
}

func (a *Vec2) Add(b *Vec2) Vec2 {
	return Vec2{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func (a *Vec2) Sub(b *Vec2) Vec2 {
	return Vec2{
		a.X - b.X,
		a.Y - b.Y,
	}
}

func (a *Vec2) Mul(m float64) Vec2 {
	return Vec2{
		X: int(float64(a.X) * m),
		Y: int(float64(a.Y) * m),
	}
}

// Returns the dot product of x and y
func (a *Vec2) Dot(b *Vec2) int {
	return a.X*b.X + a.Y*b.Y
}

// Returns the cross product of x and y
func (a *Vec2) Cross(b *Vec2) int {
	return a.X*b.Y - a.Y*b.X
}
