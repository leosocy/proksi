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

	"github.com/leosocy/proksi/pkg/loadbalancer"
	"github.com/leosocy/proksi/pkg/storage"
	"github.com/leosocy/proksi/pkg/storage/backend"

	"github.com/Sirupsen/logrus"
	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"

	"github.com/leosocy/proksi/pkg/proxy"
)

// session represents the connections between middleman and the real pxy server,
// every real pxy server corresponds to a unique session.
//
// session can carry different requests with the same pxy, and will reuse connection.
type session struct {
	pxy *proxy.Proxy
	tr  *http.Transport
}

func newSession(pxy *proxy.Proxy, tr *http.Transport) *session {
	session := &session{
		pxy: pxy,
		tr:  tr,
	}
	session.tr.Proxy = func(request *http.Request) (url *url.URL, e error) {
		return pxy.URL(), nil
	}
	return session
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

func (s *session) roundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	newContext := httptrace.WithClientTrace(req.Context(), s.newTrace(req))
	resp, err = s.tr.RoundTrip(req.WithContext(newContext))
	if err != nil {
		logrus.Warnf("Got err from RoundTrip, %+v", err)
	}
	return
}

// Weight implements the Endpoint.Weight interface.
func (s *session) Weight() int {
	return int(s.pxy.Score)
}

func (s *session) String() string {
	return s.pxy.String()
}

func (s *session) close() {
	s.tr.CloseIdleConnections()
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

type roundTripRes struct {
	s    *session
	resp *http.Response
	err  error
}

type SessionManager struct {
	lb                  loadbalancer.LoadBalancer
	pxyCh               chan *proxy.Proxy
	defaultRoundTripper http.RoundTripper
}

func NewSessionManager(nb backend.NotifyBackend, strategy loadbalancer.Strategy) *SessionManager {
	sm := &SessionManager{
		lb:    loadbalancer.NewLoadBalancer(strategy),
		pxyCh: make(chan *proxy.Proxy, 128),
	}
	nb.Attach(backend.NewInsertionWatcher(func(pxy *proxy.Proxy) {
		sm.pxyCh <- pxy
	}, storage.FilterUptime(90)))
	sm.init()
	return sm
}

func (sm *SessionManager) init() {
	go func() {
		for {
			select {
			case pxy := <-sm.pxyCh:
				session := newSession(pxy, newDefaultSessionTransport())
				sm.lb.AddEndpoint(session)
			}
		}
	}()
}

func (sm *SessionManager) pickOne() (*session, error) {
	endpoint := sm.lb.Select()
	if endpoint == nil {
		return nil, errSessionUnavailable
	}
	return endpoint.(*session), nil
}

// RoundTrip implements the goproxy.RoundTripper interface.
func (sm *SessionManager) RoundTrip(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	rtResCh := make(chan roundTripRes)
	go func() {
		var (
			session *session
			resp    *http.Response
			err     error
		)
		if session, err = sm.pickOne(); err != nil {
			resp, err = http.DefaultTransport.RoundTrip(req)
		} else {
			logrus.Infof("RoundTrip through session:%s", session.String())
			resp, err = session.roundTrip(req, ctx)
		}
		rtResCh <- roundTripRes{session, resp, err}
	}()

	select {
	case v := <-rtResCh:
		if v.err != nil {
			logrus.Warnf("Remove session:%s from load balancer", v.s.String())
			sm.lb.DelEndpoint(v.s)
		}
		return v.resp, v.err
	}
}
