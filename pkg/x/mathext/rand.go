// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math"
	"math/rand"
	"sync"
)

// Linear congruential random number generator
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

//集合内随机取数, [min,max]
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

//四舍五入
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

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}
