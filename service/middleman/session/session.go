// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package session

import (
	"github.com/Leosocy/IntelliProxy/pkg/loadbalancer"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
	"github.com/Leosocy/IntelliProxy/pkg/storage/backend"
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

// Session represents the connections between middleman and the real pxy server,
// every real pxy server corresponds to a unique Session.
//
// Session can carry different requests with the same pxy, and will reuse connection.
type Session struct {
	pxy *proxy.Proxy
	tr  *http.Transport
}

func (s *Session) newTrace(req *http.Request) *httptrace.ClientTrace {
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
func (s *Session) RoundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	newContext := httptrace.WithClientTrace(req.Context(), s.newTrace(req))
	resp, err = s.tr.RoundTrip(req.WithContext(newContext))
	if err != nil {
		logrus.Warnf("Got err from RoundTrip, %+v", err)
	}
	return
}

// Weight implements the Endpoint.Weight interface.
func (s *Session) Weight() int {
	return int(s.pxy.Score)
}

func (s *Session) String() string {
	return s.pxy.String()
}

func (s *Session) close() {
	s.tr.CloseIdleConnections()
}

var (
	ErrSessionUnavailable = errors.New("Session unavailable")
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
	sessions        []*Session
	roundRobinStart uint64
	sessionMu       sync.RWMutex
}

func (pool *sessionPool) new(pxy *proxy.Proxy, tr *http.Transport) *Session {
	session := &Session{
		pxy: pxy,
		tr:  tr,
	}
	session.tr.Proxy = func(request *http.Request) (url *url.URL, e error) {
		return url.Parse(pxy.URL())
	}
	return session
}

func (pool *sessionPool) put(s *Session) {
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

// remove remove Session from pool
func (pool *sessionPool) remove(s *Session) {
	pool.sessionMu.Lock()
	defer pool.sessionMu.Unlock()
	pool.removeSessionLocked(s)
}

// sp.sessionMu must be held
func (pool *sessionPool) removeSessionLocked(s *Session) {
	if s == nil {
		panic("nil Session")
	}
	s.close()
	for idx, v := range pool.sessions {
		if v.pxy.Equal(s.pxy) {
			copy(pool.sessions[idx:], pool.sessions[idx+1:])
			pool.sessions = pool.sessions[:len(pool.sessions)-1]
			return
		}
	}
}

type Manager struct {
	pool  *sessionPool
	lb    loadbalancer.LoadBalancer
	pxyCh chan *proxy.Proxy
}

func NewManager(nb backend.NotifyBackend, strategy loadbalancer.Strategy) *Manager {
	sm := &Manager{
		pool:  &sessionPool{},
		lb:    loadbalancer.NewLoadBalancer(strategy),
		pxyCh: make(chan *proxy.Proxy, 128),
	}
	nb.Attach(backend.NewInsertionWatcher(func(pxy *proxy.Proxy) {
		sm.pxyCh <- pxy
	}, storage.FilterScore(90)))
	sm.init()
	return sm
}

func (m *Manager) init() {
	go func() {
		for {
			select {
			case pxy := <-m.pxyCh:
				session := m.pool.new(pxy, newDefaultSessionTransport())
				m.pool.put(session)
				m.lb.AddEndpoint(session)
			}
		}
	}()
}

func (m *Manager) PickOne() (*Session, error) {
	endpoint := m.lb.Select()
	if endpoint == nil {
		return nil, ErrSessionUnavailable
	}
	return endpoint.(*Session), nil
}
