// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultMaxBytesPerFile = 1024 * 1024 * 20 // 20M
	FlushInterval          = 10               // 10ms
)

var (
	ErrCannotRotateFile = errors.New("cannot create rotate file")
)

// FileWriter with log-rotation and auto-archive
type FileWriter struct {
	done            chan struct{}
	wg              sync.WaitGroup //
	syncReqChan     chan struct{}  //
	syncRespChan    chan struct{}  //
	bus             chan []byte    // 待写内容
	maxBytesPerFile int            // 每个文件最大写入字节
	filename        string         // 写入文件名
	currentWrote    int            // 当前写入字节
}

func NewFileWriter(name string, maxBytesPerFile int) *FileWriter {
	if maxBytesPerFile <= 0 {
		maxBytesPerFile = DefaultMaxBytesPerFile
	}
	return &FileWriter{
		syncReqChan:     make(chan struct{}),
		syncRespChan:    make(chan struct{}),
		bus:             make(chan []byte, 8000),
		filename:        name,
		maxBytesPerFile: maxBytesPerFile,
	}
}

func (w *FileWriter) Init() error {
	if cap(w.done) > 0 {
		return nil
	}
	w.done = make(chan struct{}, 1)
	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if fi, err := f.Stat(); err == nil {
		w.currentWrote = int(fi.Size())
	} else {
		return err
	}

	w.wg.Add(1)
	go w.writePump()

	return nil
}

func (w *FileWriter) Flush() error {
	w.syncReqChan <- struct{}{}
	<-w.syncRespChan
	return nil
}

func (w *FileWriter) Close() {
	close(w.done)
	w.wg.Wait()
	close(w.bus)
	w.bus = nil
	w.done = nil
}

func (w *FileWriter) Write(b []byte) (int, error) {
	if w.maxBytesPerFile > 0 && len(b) > w.maxBytesPerFile {
		return 0, errors.Errorf("content size out of range %d/%d", len(b), w.maxBytesPerFile)
	}
	w.bus <- b
	return 0, nil
}

// 当my.log文件写满了后, 把my.log重命名为my.log.1, 并且把my.log.1压缩为my.log.1.tar.gz
func (w *FileWriter) rotate() error {
	var targetFilename string
	for i := 1; i <= math.MaxUint8; i++ {
		var filename = fmt.Sprintf("%s.%d", w.filename, i)
		var archiveFile = fmt.Sprintf("%s.%d.tar.gz", w.filename, i)
		if !IsFileExist(filename) && !IsFileExist(archiveFile) {
			targetFilename = filename
			break
		}
	}
	if targetFilename == "" || targetFilename == w.filename {
		return ErrCannotRotateFile
	}
	if err := os.Rename(w.filename, targetFilename); err != nil {
		return err
	}
	// 重命名文件后，保证可以使用原文件名继续写入，这里异步进行gzip压缩
	go func() {
		if err := ArchiveGzipFile(targetFilename); err != nil {
			log.Printf("ArchiveGzipFile %s: %v\n", targetFilename, err)
		} else {
			os.Remove(targetFilename)
		}
	}()
	return nil
}

func (w *FileWriter) write() error {
	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		select {
		case b, ok := <-w.bus:
			if !ok {
				return nil
			}
			n, err := f.Write(b)
			if err != nil {
				return err
			}
			w.currentWrote += n

		default:
			return nil
		}
	}
}

func (w *FileWriter) flush() {
	if err := w.write(); err != nil {
		log.Printf("%v\n", err)
		return
	}
	if w.currentWrote >= w.maxBytesPerFile {
		if err := w.rotate(); err != nil {
			log.Printf("rotate: %v\n", err)
		}
		w.currentWrote = 0
	}
}

func (w *FileWriter) sync() {
	defer func() { w.syncRespChan <- struct{}{} }()
	w.flush()
}

// 刷新线程，定时执行刷盘
func (w *FileWriter) writePump() {
	defer w.wg.Done()
	defer w.flush()
	ticker := time.NewTicker(time.Millisecond * FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.flush()

		case <-w.syncReqChan:
			w.sync()

		case <-w.done:
			return
		}
	}
}
