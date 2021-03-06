// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"net"
	"sync"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
)

type TcpServer struct {
	done    chan struct{}
	wg      sync.WaitGroup
	backlog chan fatchoy.Endpoint // queue of incoming connections
	errors  chan error            // error queue
	lns     []net.Listener        // listener list
	encoder fatchoy.ProtocolCodec // message encoding/decoding
	inbound chan *fatchoy.Packet  // incoming message buffer queue
	outsize int32                 // size of outbound message queue
}

func NewTcpServer(encoder fatchoy.ProtocolCodec, inbound chan *fatchoy.Packet, outsize int32) *TcpServer {
	return &TcpServer{
		inbound: inbound,
		encoder: encoder,
		outsize: outsize,
		done:    make(chan struct{}),
		backlog: make(chan fatchoy.Endpoint, 128),
		errors:  make(chan error, 16),
	}
}

func (s *TcpServer) BacklogChan() chan fatchoy.Endpoint {
	return s.backlog
}

func (s *TcpServer) ErrorChan() chan error {
	return s.errors
}

func (s *TcpServer) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lns = append(s.lns, ln)
	s.wg.Add(1)
	go s.serve(ln)
	return nil
}

func (s *TcpServer) checkIfExit() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *TcpServer) serve(ln net.Listener) {
	defer s.wg.Done()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("accept error: %v", err)
			// check if we should exit
			if s.checkIfExit() {
				return
			}
			return
		}

		// check if we should exit
		if s.checkIfExit() {
			return
		}

		s.accept(conn)
	}
}

func (s *TcpServer) accept(conn net.Conn) {
	var endpoint = NewTcpConn(0, conn, s.encoder, s.errors, s.inbound, s.outsize, nil)
	s.backlog <- endpoint // this may block current goroutine
}

func (s *TcpServer) Close() {
	close(s.done)
	for i, ln := range s.lns {
		ln.Close()
		s.lns[i] = nil
	}
	s.wg.Wait()
	close(s.backlog)
	close(s.errors)
	s.backlog = nil
	s.errors = nil
	s.lns = nil
	s.inbound = nil
}

func (s *TcpServer) Shutdown() {
	s.Close()
}
