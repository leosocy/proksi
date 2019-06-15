// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

// Server is a middleman between client and proxy server.
type Server struct {
	*net.Dialer
	*goproxy.ProxyHttpServer
}

func NewServer() *Server {
	ps := goproxy.NewProxyHttpServer()
	middleman := &Server{
		Dialer: &net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 120 * time.Second,
		},
		ProxyHttpServer: ps,
	}
	middleman.Tr = &http.Transport{
		Proxy:                 middleman.selectProxy,
		DialContext:           (middleman.Dialer).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	middleman.ConnectDial = middleman.httpsConnectDialToProxyHandler
	return middleman
}

func (s *Server) selectProxy(r *http.Request) (u *url.URL, err error) {
	return url.Parse("http://127.0.0.1:8888")
}

func (s *Server) httpsConnectDialToProxyHandler(network, addr string) (net.Conn, error) {
	proxyURL, err := s.selectProxy(nil)
	if err != nil {
		return nil, err
	}
	if strings.IndexRune(proxyURL.Host, ':') == -1 {
		proxyURL.Host += ":80"
	}
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	c, err := s.Dialer.Dial(network, proxyURL.Host)
	if err != nil {
		return nil, err
	}
	connectReq.Write(c)
	// Read response.
	// Okay to use and discard buffered reader here, because
	// TLS server will not speak until spoken to.
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		c.Close()
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		c.Close()
		return nil, errors.New("proxy refused connection" + string(resp))
	}
	return c, nil
}
