// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"net/http"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"

	"github.com/Leosocy/IntelliProxy/pkg/storage"
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
	sessionP *sessionPool
	*goproxy.ProxyHttpServer
}

func NewServer(storage storage.Storage) *Server {
	s := &Server{
		sessionP:        &sessionPool{},
		ProxyHttpServer: goproxy.NewProxyHttpServer(),
	}
	s.Verbose = true
	pxy, _ := proxy.NewProxy("47.94.135.32", "8118")
	session := s.sessionP.new(pxy, newDefaultSessionTransport())
	s.sessionP.put(session)
	s.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	// TODO:
	//  1. 选择合适的session，并且设置ctx.RoundTripper
	s.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (request *http.Request, response *http.Response) {
		ctx.RoundTripper, _ = s.sessionP.get()
		return req, nil
	})
	return s
}
