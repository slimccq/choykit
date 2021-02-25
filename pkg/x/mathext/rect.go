// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

type Rectangle struct {
	X, Y          int // 原点
	Width, Height int // 宽度、高度
}

var (
	EmptyRect Rectangle
)

func NewRectangle(x, y, w, h int) *Rectangle {
	return &Rectangle{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	}
}

func (r *Rectangle) GetVertexes() [4]Point2 {
	return [4]Point2{
		{r.X, r.Y},
		{r.X+r.Width, r.Y},
		{r.X+r.Width, r.Y+r.Height},
		{r.X, r.Y+r.Height},
	}
}

// Inflates this rectangle
func (r *Rectangle) Inflate(width, height int) {
	r.X -= width
	r.Y -= height
	r.Width += 2 * width
	r.Height += 2 * height
}

// Determines if a point is contained within the rectangle
func (r *Rectangle) Contains(x, y int) bool {
	return r.X <= x && x < r.X+r.Width &&
		r.Y <= y && y < r.Y+r.Height
}

func (r *Rectangle) ContainsPoint(pt Point2) bool {
	return r.Contains(pt.X, pt.Y)
}

func (r *Rectangle) ContainsRegion(rec *Rectangle) bool {
	return r.X <= rec.X && (rec.X+rec.Width) <= (r.X+r.Width) &&
		r.Y <= rec.Y && (rec.Y+rec.Height) <= (r.Y+r.Height)
}

// Determines if this rectangle intersets with rect.
func (r *Rectangle) IsIntersectsWith(rec *Rectangle) bool {
	return (rec.X < r.X+r.Width) && r.X < (rec.X+rec.Width) &&
		(rec.Y < r.Y+r.Height) && r.Y < (rec.Y+rec.Height)
}

// Creates a rectangle that represents the intersetion between a and b
func RectIntersect(a *Rectangle, b *Rectangle) *Rectangle {
	var x1 = Int.Max(a.X, b.X)
	var x2 = Int.Max(a.X+a.Width, b.X+b.Width)
	var y1 = Int.Max(a.Y, b.Y)
	var y2 = Int.Max(a.Y+a.Height, b.Y+b.Height)
	if x2 >= x1 && y2 >= y1 {
		return NewRectangle(x1, y1, x2-x1, y2-y1)
	}
	return &EmptyRect
}

// Creates a rectangle that represents the union between a and b
func RectUnion(a *Rectangle, b *Rectangle) *Rectangle {
	var x1 = Int.Max(a.X, b.X)
	var x2 = Int.Max(a.X+a.Width, b.X+b.Width)
	var y1 = Int.Max(a.Y, b.Y)
	var y2 = Int.Max(a.Y+a.Height, b.Y+b.Height)
	return NewRectangle(x1, y1, x2-x1, y2-y1)
}
