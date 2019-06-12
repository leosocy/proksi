// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
	"github.com/Sirupsen/logrus"
	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
)

// session represents the connections between middleman and the real pxy server,
// every real pxy server corresponds to a unique session.
//
// session can carry different requests with the same pxy, and will reuse connection.
type session struct {
	pxy *proxy.Proxy
	tr  *http.Transport
}

func (s *session) newTrace(req *http.Request) *httptrace.ClientTrace {
	tracer := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			logrus.Infof("GotConn: %+v", info)
		},
		PutIdleConn: func(err error) {
			logrus.Infof("PutIdleConn: %+v", err)
		},
	}
	return tracer
}

// RoundTrip implements the goproxy.RoundTripper interface.
func (s *session) RoundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	newContext := httptrace.WithClientTrace(req.Context(), s.newTrace(req))
	resp, err = s.tr.RoundTrip(req.WithContext(newContext))
	if err != nil {
		logrus.Warnf("Got err from RoundTrip, %+v", err)
	}
	return
}

func (s *session) close() {
	s.tr.CloseIdleConnections()
}

type reqKey struct {
	scheme, host string
}

func (k reqKey) String() string {
	return fmt.Sprintf("%s|%s", k.scheme, k.host)
}

var (
	errSessionUnavailable = errors.New("session unavailable")
)

func newDefaultSessionTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   8 * time.Second,
			KeepAlive: 16 * time.Second,
		}).DialContext,
		MaxIdleConns:          64,
		MaxConnsPerHost:       8,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       8 * time.Minute,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 4 * time.Second,
	}
}

type sessionPool struct {
	sessions  []*session
	sessionMu sync.RWMutex
}

func (pool *sessionPool) new(pxy *proxy.Proxy, tr *http.Transport) *session {
	session := &session{
		pxy: pxy,
		tr:  tr,
	}
	session.tr.Proxy = func(request *http.Request) (url *url.URL, e error) {
		return url.Parse(pxy.URL())
	}
	return session
}

func (pool *sessionPool) put(s *session) {
	if s == nil {
		return
	}
	pool.sessionMu.Lock()
	defer pool.sessionMu.Unlock()
	for _, exist := range pool.sessions {
		if exist.pxy.Equal(s.pxy) {
			return
		}
	}
	pool.sessions = append(pool.sessions, s)
}

func (pool *sessionPool) get() (*session, error) {
	// TODO: 根据策略选择
	pool.sessionMu.RLock()
	defer pool.sessionMu.RUnlock()
	if len(pool.sessions) == 0 {
		return nil, errSessionUnavailable
	}
	return pool.sessions[0], nil
}

// remove remove session from pool
func (pool *sessionPool) remove(s *session) {
	pool.sessionMu.Lock()
	defer pool.sessionMu.Unlock()
	pool.removeSessionLocked(s)
}

// sp.sessionMu must be held
func (pool *sessionPool) removeSessionLocked(s *session) {
	s.close()
	sessions := pool.sessions
	switch len(sessions) {
	case 0:
		// do nothing
	case 1:
		pool.sessions = []*session{}
	default:
		for i, v := range sessions {
			if !v.pxy.Equal(s.pxy) {
				continue
			}
			copy(sessions[i:], sessions[i+1:])
			pool.sessions = sessions[:len(sessions)-1]
			break
		}
	}
}
