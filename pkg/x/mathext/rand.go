// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
	"math/rand"
	"sync"
)

// 线性同余数法 Linear congruential random number generator
// see https://en.wikipedia.org/wiki/Linear_congruential_generator
type LCRNG struct {
	seed  uint32
	guard sync.Mutex
}

func (g *LCRNG) Seed(seed uint32) {
	g.guard.Lock()
	g.seed = seed*214013 + 2531011
	g.guard.Unlock()
}

func (g *LCRNG) Rand() uint32 {
	g.guard.Lock()
	g.seed = g.seed*214013 + 2531011
	var r = uint32(g.seed>>16) & 0x7fff
	g.guard.Unlock()
	return r
}

// Random integer in [min, max]
func RandInt(min, max int) int {
	if min > max {
		panic("RandInt,min greater than max")
	}
	if min == max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

// Random number in [min, max]
func RandFloat(min, max float64) float64 {
	if min > max {
		panic("RandFloat: min greater than max")
	}
	if min == max {
		return min
	}
	return rand.Float64()*(max-min) + min
}

// 集合内随机取数, [min,max]
func RangePerm(min, max int) []int {
	if min > max {
		panic("RangePerm: min greater than max")
	}
	if min == max {
		return []int{min}
	}
	list := rand.Perm(max - min + 1)
	for i := 0; i < len(list); i++ {
		list[i] += min
	}
	return list
}

// 四舍五入
func RoundHalf(v float64) int {
	return int(RoundFloat(v))
}

// https://github.com/montanaflynn/stats/blob/master/round.go
func RoundFloat(x float64) float64 {
	// If the float is not a number
	if math.IsNaN(x) {
		return math.NaN()
	}

	// Find out the actual sign and correct the input for later
	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	// Get the actual decimal number as a fraction to be compared
	_, decimal := math.Modf(x)

	// If the decimal is less than .5 we round down otherwise up
	var rounded float64
	if decimal >= 0.5 {
		rounded = math.Ceil(x)
	} else {
		rounded = math.Floor(x)
	}

	// Finally we do the math to actually create a rounded number
	return rounded * sign
}
