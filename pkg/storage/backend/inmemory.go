// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package backend

import (
	"hash/fnv"
	"net"
	"sync"

	"github.com/Leosocy/IntelliProxy/pkg/storage"

	"github.com/HuKeping/rbtree"
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

type comparableProxy struct {
	pxy *proxy.Proxy
}

// Less implements rbtree.Less method.
func (p *comparableProxy) Less(than rbtree.Item) bool {
	thanP := than.(*comparableProxy)
	if p.pxy.IP.Equal(thanP.pxy.IP) {
		return false
	}
	return p.pxy.Score <= thanP.pxy.Score
}

// InMemoryBackend is a simple local in memory backend.
type InMemoryBackend struct {
	m    map[uint64]*proxy.Proxy // map[hash64(IP)]proxy
	rbt  *rbtree.Rbtree
	lock sync.RWMutex
}

// NewInMemoryBackend returns new InMemoryBackend with default configurations.
func NewInMemoryBackend() *InMemoryBackend {
	return &InMemoryBackend{
		m:   make(map[uint64]*proxy.Proxy),
		rbt: rbtree.New(),
	}
}

func (s *InMemoryBackend) insert(p *proxy.Proxy) error {
	hasher := fnv.New64()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rbt.Insert(&comparableProxy{pxy: p})
	if _, err := hasher.Write(p.IP); err != nil {
		return err
	}
	s.m[hasher.Sum64()] = p
	return nil
}

func (s *InMemoryBackend) Insert(p *proxy.Proxy) error {
	if p == nil || p.Score <= 0 {
		return ErrProxyInvalid
	}
	if sp := s.Search(p.IP); sp != nil {
		return ErrProxyDuplicated
	}
	return s.insert(p)
}

func (s *InMemoryBackend) Select(opts ...storage.SelectOption) ([]*proxy.Proxy, error) {
	sopts := storage.SelectOptions{}
	for _, opt := range opts {
		opt(&sopts)
	}
	proxies := s.TopK(0)
	for _, filter := range sopts.Filters {
		proxies = filter(proxies)
	}
	if len(proxies) == 0 || sopts.Offset >= len(proxies) {
		return nil, ErrProxyNoneAvailable
	}
	remained := len(proxies) - sopts.Offset
	if sopts.Limit == 0 || sopts.Limit >= remained {
		return proxies[sopts.Offset:], nil
	}
	return proxies[sopts.Offset : sopts.Offset+sopts.Limit], nil
}

func (s *InMemoryBackend) Search(ip net.IP) *proxy.Proxy {
	hasher := fnv.New64()
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, err := hasher.Write(ip); err == nil {
		return s.m[hasher.Sum64()]
	}
	return nil
}

func (s *InMemoryBackend) delete(p *proxy.Proxy) error {
	hasher := fnv.New64()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rbt.Delete(&comparableProxy{pxy: p})
	if _, err := hasher.Write(p.IP); err != nil {
		return err
	}
	delete(s.m, hasher.Sum64())
	return nil
}

func (s *InMemoryBackend) Delete(ip net.IP) error {
	var sp *proxy.Proxy
	if sp = s.Search(ip); sp == nil {
		return ErrProxyDoesNotExists
	}
	return s.delete(sp)
}

func (s *InMemoryBackend) Update(newP *proxy.Proxy) error {
	if err := s.Delete(newP.IP); err != nil {
		return err
	}
	return s.Insert(newP)
}

func (s *InMemoryBackend) InsertOrUpdate(p *proxy.Proxy) (bool, error) {
	err := s.Insert(p)
	switch err {
	case ErrProxyDuplicated:
		return false, s.Update(p)
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (s *InMemoryBackend) Len() uint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.rbt.Len()
}

func (s *InMemoryBackend) TopK(k int) []*proxy.Proxy {
	proxies := make([]*proxy.Proxy, 0)
	s.Iter(func(pxy *proxy.Proxy) bool {
		if k == 0 || len(proxies) < k {
			proxies = append(proxies, pxy)
			return true
		}
		return false
	})
	return proxies
}

func (s *InMemoryBackend) Iter(iter Iterator) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.rbt.Descend(s.rbt.Max(), func(item rbtree.Item) bool {
		pxy := item.(*comparableProxy).pxy
		return iter(pxy)
	})
}
