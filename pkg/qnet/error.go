// Copyright © 2018-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package qnet

import (
	"errors"
	"fmt"

	"devpkg.work/choykit/pkg"
)

var (
	ErrConnIsClosing        = errors.New("connection is closing when sending")
	ErrConnOutboundOverflow = errors.New("connection outbound queue overflow")
	ErrConnForceClose       = errors.New("connection forced to close")
)

type Error struct {
	Err      error
	Endpoint choykit.Endpoint
}

func NewError(err error, endpoint choykit.Endpoint) error {
	return &Error{
		Err:      err,
		Endpoint: endpoint,
	}
}

func (e Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("node %v(%s) EOF", e.Endpoint.NodeID(), e.Endpoint.RemoteAddr())
	}
	return fmt.Sprintf("node %v(%s) %s", e.Endpoint.NodeID(), e.Endpoint.RemoteAddr(), e.Err.Error())
}
