// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/Leosocy/IntelliProxy/pkg/loadbalancer"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
	"github.com/Leosocy/IntelliProxy/pkg/storage/backend"

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

func NewSession(pxy *proxy.Proxy, tr *http.Transport) *Session {
	session := &Session{
		pxy: pxy,
		tr:  tr,
	}
	session.tr.Proxy = func(request *http.Request) (url *url.URL, e error) {
		return url.Parse(pxy.URL())
	}
	return session
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

func (s *Session) roundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
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

type SessionManager struct {
	lb                  loadbalancer.LoadBalancer
	defaultRoundTripper http.RoundTripper
	pxyCh               chan *proxy.Proxy
}

func NewSessionManager(nb backend.NotifyBackend, strategy loadbalancer.Strategy) *SessionManager {
	sm := &SessionManager{
		lb:    loadbalancer.NewLoadBalancer(strategy),
		pxyCh: make(chan *proxy.Proxy, 128),
	}
	nb.Attach(backend.NewInsertionWatcher(func(pxy *proxy.Proxy) {
		sm.pxyCh <- pxy
	}, storage.FilterScore(90)))
	sm.init()
	return sm
}

func (sm *SessionManager) init() {
	go func() {
		for {
			select {
			case pxy := <-sm.pxyCh:
				session := NewSession(pxy, newDefaultSessionTransport())
				sm.lb.AddEndpoint(session)
			}
		}
	}()
}

func (sm *SessionManager) pickOne() (*Session, error) {
	endpoint := sm.lb.Select()
	if endpoint == nil {
		return nil, ErrSessionUnavailable
	}
	return endpoint.(*Session), nil
}

// RoundTrip implements the goproxy.RoundTripper interface.
func (sm *SessionManager) RoundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	var session *Session
	if session, err = sm.pickOne(); err != nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	logrus.Infof("RoundTrip through session:%s", session.String())
	if resp, err = session.roundTrip(req, ctx); err != nil {
		logrus.Warnf("Remove session:%s from load balancer", session.String())
		sm.lb.DelEndpoint(session)
	}
	return
}
