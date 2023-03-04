// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"net/http"

	"github.com/leosocy/proksi/pkg/loadbalancer"
	"github.com/leosocy/proksi/pkg/storage/backend"

	"github.com/elazarl/goproxy"
)

// Server is a middleman between client and real pxy server.
// It run as a https server which always eavesdrop https connections,
// the purpose is to reuse the connection between middleman and the pxy server,
// avoiding TLS handshakes for every request.
//
// And, this is safe because the middleman server is usually deployed
// as a sidecar with crawler program together.
type Server struct {
	sm *SessionManager
	*goproxy.ProxyHttpServer
}

func NewServer(nb backend.NotifyBackend) *Server {
	s := &Server{
		sm:              NewSessionManager(nb, loadbalancer.WeightedRoundRobin),
		ProxyHttpServer: goproxy.NewProxyHttpServer(),
	}
	s.Verbose = true
	s.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	// TODO: 设置Transport.Proxy，包装Transport.RoundTrip 处理err实现。不要重写整个RoundTripper
	s.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (request *http.Request, response *http.Response) {
		ctx.RoundTripper = s.sm
		return req, nil
	})
	return s
}
