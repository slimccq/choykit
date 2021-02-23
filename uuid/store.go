// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

type Storage interface {
	Init() error
	Next() (int64, error)
	MustNext() int64
	Close()
}
