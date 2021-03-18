// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

// 圆形
type Circle struct {
	Center Point // 中心点
	Radius int   // 半径
}

func NewCircle(center Point, radius int) *Circle {
	return &Circle{
		Center: center,
		Radius: radius,
	}
}

// 获取包围矩形
func (c *Circle) GetSurroundRect() Rectangle {
	return Rectangle{
		Point: Point{
			X: c.Center.X - c.Radius,
			Y: c.Center.Y - c.Radius,
		},
		Width:  c.Radius * 2,
		Height: c.Radius * 2,
	}
}
