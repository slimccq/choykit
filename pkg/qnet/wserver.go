// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"context"
	"net/http"
	"time"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"github.com/gorilla/websocket"
)

// Websocket server
type WsServer struct {
	server   *http.Server
	upgrader *websocket.Upgrader   //
	pending  chan *WsConn          //
	errChan  chan error            //
	inbound  chan *fatchoy.Packet  // incoming message queue
	encoder  fatchoy.ProtocolCodec // message codec
	outsize  int                   // outgoing queue size
}

func NewWebsocketServer(addr, path string, encoder fatchoy.ProtocolCodec, inbound chan *fatchoy.Packet, outsize int) *WsServer {
	mux := http.NewServeMux()
	var server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    4096,
	}
	ws := &WsServer{
		server:  server,
		encoder: encoder,
		inbound: inbound,
		outsize: outsize,
		errChan: make(chan error, 32),
		pending: make(chan *WsConn, 128),
		upgrader: &websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
		},
	}
	mux.HandleFunc(path, ws.onRequest)
	return ws
}

func (s *WsServer) onRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade %s, %v", r.RemoteAddr, err)
		return
	}
	wsconn := NewWsConn(0, conn, s.encoder, s.errChan, s.inbound, s.outsize, nil)
	log.Infof("websocket connection %s established", wsconn.RemoteAddr())
	defer wsconn.Close()
	wsconn.Go(true, false)
	wsconn.readLoop()
}

func (s *WsServer) BacklogChan() chan *WsConn {
	return s.pending
}

func (s *WsServer) ErrChan() chan error {
	return s.errChan
}

func (s *WsServer) Go() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Errorf("ListenAndServe: %v", err)
		}
	}()
}

func (s *WsServer) Shutdown() {
	s.server.Shutdown(context.Background())
	close(s.pending)
	close(s.errChan)
	s.errChan = nil
	s.pending = nil
	s.inbound = nil
	s.server = nil
	s.encoder = nil
}
