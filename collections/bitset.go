// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"math/bits"
	"strings"
)

// 定长bitset
type BitSet struct {
	bitsize int
	bits    []uint64
}

func BitSetFrom(bitsize int, array []uint64) *BitSet {
	return &BitSet{
		bitsize: bitsize,
		bits:    array,
	}
}

func NewBitSet(bitsize int) *BitSet {
	var n = bitsize / 64
	if bitsize%64 > 0 {
		n++
	}
	return &BitSet{
		bitsize: bitsize,
		bits:    make([]uint64, n),
	}
}

func (bs *BitSet) Size() int {
	return bs.bitsize
}

// set 1 to bits[i]
func (bs *BitSet) Set(i int) bool {
	if i >= 0 && i < bs.bitsize {
		var v = uint64(1) << (i % 64)
		bs.bits[i/64] |= v
		return true
	}
	return false
}

func (bs *BitSet) Flip(i int) bool {
	if i >= 0 && i < bs.bitsize {
		bs.bits[i/64] ^= 1 << (i % 64)
		return true
	}
	return false
}

// set 0 to bits[i]
func (bs *BitSet) Clear(i int) bool {
	if i >= 0 && i < bs.bitsize {
		var v = uint64(1) << (i % 64)
		bs.bits[i/64] &= ^v
		return true
	}
	return false
}

// test bits[i]
func (bs *BitSet) Test(i int) bool {
	if i >= 0 && i < bs.bitsize {
		return bs.bits[i/64]&(1<<(i%64)) != 0
	}
	return false
}

func (bs *BitSet) ClearAll() {
	for i := 0; i < len(bs.bits); i++ {
		bs.bits[i] = 0
	}
}

func (bs *BitSet) Count() int {
	var count = 0
	for i := 0; i < len(bs.bits); i++ {
		if bs.bits[i] > 0 {
			count += bits.OnesCount64(bs.bits[i])
		}
	}
	return count
}

func (bs BitSet) String() string {
	var sb strings.Builder
	sb.Grow(bs.bitsize)
	for i := 0; i < bs.bitsize; i++ {
		if bs.Test(i) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}
