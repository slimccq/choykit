// Copyright © 2017-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package codec

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
)

type CompressType int

const (
	ZLIB  CompressType = 1
	FLATE CompressType = 2
	GZIP  CompressType = 3

	DefaultCompression = 0
	BestSpeed          = 1
	BestCompression    = 2
)

var ErrUnsupportedCodec = errors.New("unsupported codec type")

func CompressBytes(ctype CompressType, level int, data []byte) ([]byte, error) {
	var err error
	var w io.WriteCloser
	var lvl = flate.DefaultCompression
	switch level {
	case BestSpeed:
		lvl = flate.BestSpeed
	case BestCompression:
		lvl = flate.BestCompression
	}
	var buf = &bytes.Buffer{}
	switch ctype {
	case ZLIB:
		w, err = zlib.NewWriterLevel(buf, lvl)
	case FLATE:
		w, err = flate.NewWriter(buf, lvl)
	case GZIP:
		w, err = gzip.NewWriterLevel(buf, lvl)
	default:
		err = ErrUnsupportedCodec
	}
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnCompressBytes(ctype CompressType, data []byte) ([]byte, error) {
	var err error
	var r io.ReadCloser
	switch ctype {
	case ZLIB:
		r, err = zlib.NewReader(bytes.NewReader(data))
	case FLATE:
		r = flate.NewReader(bytes.NewReader(data))
	case GZIP:
		r, err = gzip.NewReader(bytes.NewReader(data))
	default:
		err = ErrUnsupportedCodec
	}
	if err != nil {
		return nil, err
	}
	var buf = &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		r.Close()
		return nil, err
	}
	if err = r.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
