// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package gateway

import (
	"net/http"
	"net/url"
	"time"

	"devpkg.work/choykit/pkg/codec"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/qnet"
	"github.com/gorilla/websocket"
)

// websocket server
type WsServer struct {
	*http.Server
	url      string                // Websocket URL地址
	upgrader *websocket.Upgrader   //
	encoder  fatchoy.ProtocolCodec //
}

func (s *WsServer) Start() {
	log.Infof("serve websocket client at %s", s.url)
	go s.ListenAndServe()
}

func (g *Service) createWSServer(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	server := &WsServer{
		url: rawurl,
		Server: &http.Server{
			Addr:              u.Host,
			Handler:           mux,
			ReadTimeout:       100 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		upgrader: &websocket.Upgrader{
			HandshakeTimeout: 5 * time.Second,
		},
	}
	server.encoder = codec.NewServerProtocolCodec()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) { g.handleWebRequest(server, w, r) })
	g.wservers = append(g.wservers, server)
	return nil
}

func (g *Service) handleWebRequest(server *WsServer, w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade %s, %v", r.RemoteAddr, err)
		return
	}
	var node = g.nextSession()
	if count := g.sessions.Size() + 1; count > int(g.pcu) {
		g.pcu = uint32(count) // dirty write is OK
	}
	var sid = NodeToSession(node)
	log.Infof("Websocket client #%d connected, %v", sid, conn.RemoteAddr())

	env := g.Context().Env()
	session := qnet.NewWsConn(node, conn, server.encoder, nil, nil,
		env.EndpointOutboundQueueSize, g.cStats)
	defer g.closeSession(session)
	session.SetContext(g.Context())

	// TODO: handshaking

	for !g.IsClosing() {
		pkt := fatchoy.MakePacket()
		if err := session.ReadPacket(pkt); err != nil {
			break
		}
		// client的消息只能指定本区服的backend
		pkt.Endpoint = session
		g.dispatchPacket(pkt)
	}
}
