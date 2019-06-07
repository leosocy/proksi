// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"bufio"
	"errors"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type connectDialFunc func(network, address string) (net.Conn, error)

type proxyURLGetter func() (*url.URL, error)

// Server is a middleman between client and proxy server.
type Server struct {
	storage storage.Storage
	*goproxy.ProxyHttpServer
}

func httpConnectDialToProxyHandler(pg proxyURLGetter) func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		return pg()
	}
}

func httpsConnectDialToProxyHandler(pg proxyURLGetter, dial connectDialFunc) connectDialFunc {
	return func(network, addr string) (net.Conn, error) {
		proxyURL, err := pg()
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
		c, err := dial(network, proxyURL.Host)
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
}

func NewServer(storage storage.Storage) *Server {
	ps := goproxy.NewProxyHttpServer()
	middleman := &Server{
		storage:         storage,
		ProxyHttpServer: ps,
	}
	rand.Seed(time.Now().UnixNano())
	middleman.Tr.Proxy = httpConnectDialToProxyHandler(middleman.getProxyURL)
	middleman.ConnectDial = httpsConnectDialToProxyHandler(middleman.getProxyURL, net.Dial)
	return middleman
}

func (s *Server) getProxyURL() (*url.URL, error) {
	proxies := s.storage.TopK(20)
	if len(proxies) == 0 {
		return nil, errors.New("no proxy available")
	}
	index := rand.Int() % len(proxies)
	return url.Parse(proxies[index].URL())
}
