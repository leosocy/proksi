// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package selector

import (
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// TODO(leosocy): Selector可以是一个interface，现有实现其实是一个基于内存的PooledSelector，
type Selector struct {
	Strategy  Strategy
	Storage   storage.Storage
	proxyPool []*proxy.Proxy
}

// Next is a function that returns the next proxy
// based on the selector's strategy
type Next func() (*proxy.Proxy, error)

// Filter is used to filter a proxy during the selection process
type Filter func([]*proxy.Proxy) []*proxy.Proxy

// Strategy is a selection strategy e.g random, round robin
type Strategy func([]*proxy.Proxy) Next

var (
	ErrNoneAvailable = errors.New("none available")
)

// Select returns a function which should return the next proxy
func (s *Selector) Select(opts ...SelectOption) (Next, error) {
	sopts := SelectOptions{Strategy: s.Strategy}
	for _, opt := range opts {
		opt(&sopts)
	}

	// get top k(pool size) proxies
	proxies := s.Storage.TopK(0)
	for _, filter := range sopts.Filters {
		proxies = filter(proxies)
	}

	if len(proxies) == 0 {
		return nil, ErrNoneAvailable
	}
	return sopts.Strategy(proxies), nil
}

// Mark sets the proxy value to 0 when the error occurs
func (s *Selector) Mark(pxy *proxy.Proxy, err error) {
	if err != nil {
		pxy.AddScore(-proxy.MaximumScore)
		if err := s.Storage.Update(pxy); err == nil {
			logrus.Infof("err occurs when use %s, now set it's score to 0", pxy.URL())
		}
	}
}

// Reset reset pool back to zero
func (s *Selector) Reset() {
	s.proxyPool = s.proxyPool[:0]
}
