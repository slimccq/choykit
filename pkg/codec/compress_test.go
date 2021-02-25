// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package codec

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"devpkg.work/choykit/pkg/x/strutil"
)

var showCompressRate = false

func testCase(t *testing.T, typ CompressType, fn func(int) []byte) {
	var size = 1
	for size < 10000 {
		text := fn(size)
		size += rand.Int() % 100
		compressed, err := CompressBytes(typ, DefaultCompression, text)
		if err != nil {
			t.Fatalf("compress: %v", err)
		}
		uncompressed, err := UnCompressBytes(typ, compressed)
		if err != nil {
			t.Fatalf("compress: %v", err)
		}
		if !bytes.Equal(text, uncompressed) {
			t.Fatalf("not equal")
		}
		if showCompressRate {
			var before, after = len(text), len(compressed)
			var rate = float64(after) / float64(before)
			if rate < 1.0 {
				t.Logf("compress %v rate: %d/%d = %f", typ, before, after, rate)
			}
		}
	}
}

func TestCompressZlib(t *testing.T) {
	testCase(t, ZLIB, strutil.RandBytes)
}

func TestCompressFlate(t *testing.T) {
	testCase(t, FLATE, strutil.RandBytes)
}

func TestCompressGzip(t *testing.T) {
	testCase(t, GZIP, strutil.RandBytes)
}

func TestCompressWriter(t *testing.T) {
	for _, typ := range []CompressType{ZLIB, FLATE, GZIP} {
		for i := 1; i <= 100; i++ {
			var data = strutil.RandBytes(256 + rand.Int()%1024)
			compressed, err := CompressBytes(typ, DefaultCompression, data)
			if err != nil {
				t.Fatalf("doCompressWithWriter: %v, turn: %d", err, i)
			}
			uncompressed, err := UnCompressBytes(typ, compressed)
			if err != nil {
				t.Fatalf("compress: %v, turn: %d", err, i)
			}
			if !bytes.Equal(data, uncompressed) {
				t.Fatalf("not equal")
			}
		}
	}
}

func TestCompressRate(t *testing.T) {
	for _, lvl := range []int{DefaultCompression, BestCompression, BestSpeed} {
		var stats = make([]float64, 0, 100)
		for i := 1; i <= 100; i++ {
			var data = strutil.RandBytes(64 + rand.Int()%8192)
			compressed, err := CompressBytes(ZLIB, lvl, data)
			if err != nil {
				t.Fatalf("doCompressWithWriter: %v, turn: %d", err, i)
			}
			var rate = float64(len(compressed)) / float64(len(data))
			stats = append(stats, rate)
		}
		var sum float64
		for _, v := range stats {
			sum += v
		}
		var avg = sum / float64(len(stats))
		fmt.Printf("compress level %v, total %v, avg: %v\n", lvl, sum, avg)
	}
}

func benchCompress(b *testing.B, typ CompressType, level int) {
	for i := 0; i < b.N; i++ {
		var data = strutil.RandBytes(8192)
		_, err := CompressBytes(typ, level, data)
		if err != nil {
			b.Fatalf("compress: %v", err)
		}
	}
	b.StopTimer()
}

func BenchmarkZlibCompress1(b *testing.B) {
	benchCompress(b, ZLIB, DefaultCompression)

}

func BenchmarkZlibCompress2(b *testing.B) {
	benchCompress(b, ZLIB, BestCompression)
}

func BenchmarkZlibCompress3(b *testing.B) {
	benchCompress(b, ZLIB, BestSpeed)
}

func BenchmarkFlateCompress(b *testing.B) {
	benchCompress(b, FLATE, DefaultCompression)
}

func BenchmarkGzipCompress(b *testing.B) {
	benchCompress(b, GZIP, DefaultCompression)
}

func benchUncompress(b *testing.B, typ CompressType) {
	b.StopTimer()

	var data = strutil.RandBytes(1024)
	compressed, err := CompressBytes(typ, DefaultCompression, data)
	if err != nil {
		b.Fatalf("compress: %v", err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := UnCompressBytes(typ, compressed)
		if err != nil {
			b.Fatalf("compress: %v", err)
		}
		b.StopTimer()
		b.StartTimer()
	}
}

func BenchmarkZlibUncompress(b *testing.B) {
	benchUncompress(b, ZLIB)
}

func BenchmarkInflateUncompress(b *testing.B) {
	benchUncompress(b, FLATE)
}

func BenchmarkGzipUncompress(b *testing.B) {
	benchUncompress(b, GZIP)
}
