// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"github.com/Sirupsen/logrus"
	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// session represents the connections between middleman and the real proxy server,
// every real proxy server corresponds to a unique session.
//
// session can carry different requests with the same proxy, and will reuse connection.
type session struct {
	id     int64
	rawurl string
	tr     *http.Transport
	ctxs   map[int64]*goproxy.ProxyCtx
	traces map[int64]*httptrace.ClientTrace
}

func (s *session) newTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			logrus.Infof("GetConn: %s", hostPort)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			logrus.Infof("GotConn: %+v", info)
		},
		PutIdleConn: func(err error) {
			logrus.Infof("PutIdleConn: %+v", err)
		},
	}
}

// RoundTrip implements the goproxy.RoundTripper interface.
func (s *session) RoundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	newContext := httptrace.WithClientTrace(req.Context(), s.newTrace())
	return s.tr.RoundTrip(req.WithContext(newContext))
}

var (
	errSessionUnavailable = errors.New("session unavailable")
)

func newDefaultSessionTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 8 * time.Second,
		}).DialContext,
		MaxIdleConns:        64,
		MaxConnsPerHost:     8,
		MaxIdleConnsPerHost: 8,
		IdleConnTimeout:     8 * time.Minute,
		TLSHandshakeTimeout: 4 * time.Second,
	}
}

type sessionManager struct {
	startID   int64
	sessions  []*session
	sessionMu sync.RWMutex
}

func (sm *sessionManager) newSession(rawurl string, tr *http.Transport) {
	session := &session{
		id:     atomic.AddInt64(&sm.startID, 1),
		rawurl: rawurl,
		tr:     tr,
		ctxs:   make(map[int64]*goproxy.ProxyCtx),
		traces: make(map[int64]*httptrace.ClientTrace),
	}
	session.tr.Proxy = func(request *http.Request) (url *url.URL, e error) {
		return url.Parse(session.rawurl)
	}
	sm.putSession(session)
}

func (sm *sessionManager) putSession(s *session) {
	sm.sessionMu.Lock()
	defer sm.sessionMu.Unlock()
	for _, exist := range sm.sessions {
		if exist.id == s.id {
			logrus.Fatalf("dup session %p", s)
		}
	}
	sm.sessions = append(sm.sessions, s)
}

func (sm *sessionManager) getSession() (*session, error) {
	// TODO: 根据策略选择
	sm.sessionMu.RLock()
	defer sm.sessionMu.RUnlock()
	if len(sm.sessions) == 0 {
		return nil, errSessionUnavailable
	}
	return sm.sessions[0], nil
}
