// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"hash/fnv"
	"net"
	"sync"

	"github.com/HuKeping/rbtree"
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

type compareableProxy struct {
	pxy *proxy.Proxy
}

// Less implements rbtree.Less method.
func (p *compareableProxy) Less(than rbtree.Item) bool {
	thanP := than.(*compareableProxy)
	if p.pxy.IP.Equal(thanP.pxy.IP) {
		return false
	}
	return p.pxy.Score <= thanP.pxy.Score
}

// InMemoryStorage is a simple storage.
type InMemoryStorage struct {
	m    map[uint64]*proxy.Proxy // map[hash64(IP)]proxy
	rbt  *rbtree.Rbtree
	lock sync.RWMutex
}

// NewInMemoryStorage returns new InMemoryStorage with default configurations.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		m:   make(map[uint64]*proxy.Proxy),
		rbt: rbtree.New(),
	}
}

func (s *InMemoryStorage) insert(p *proxy.Proxy) {
	hasher := fnv.New64()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rbt.Insert(&compareableProxy{pxy: p})
	hasher.Write(p.IP)
	s.m[hasher.Sum64()] = p
}

func (s *InMemoryStorage) Insert(p *proxy.Proxy) error {
	if p == nil || p.Score <= 0 {
		return ErrProxyInvalid
	}
	if sp := s.Search(p.IP); sp != nil {
		return ErrProxyDuplicated
	}
	s.insert(p)
	return nil
}

func (s *InMemoryStorage) Search(ip net.IP) *proxy.Proxy {
	hasher := fnv.New64()
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, err := hasher.Write(ip); err == nil {
		return s.m[hasher.Sum64()]
	}
	return nil
}

func (s *InMemoryStorage) delete(p *proxy.Proxy) {
	hasher := fnv.New64()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rbt.Delete(&compareableProxy{pxy: p})
	hasher.Write(p.IP)
	delete(s.m, hasher.Sum64())
}

func (s *InMemoryStorage) Delete(ip net.IP) error {
	var sp *proxy.Proxy
	if sp = s.Search(ip); sp == nil {
		return ErrProxyDoesNotExists
	}
	s.delete(sp)
	return nil
}

// Update implements storage.Update.
func (s *InMemoryStorage) Update(newP *proxy.Proxy) error {
	if err := s.Delete(newP.IP); err != nil {
		return err
	}
	return s.Insert(newP)
}

func (s *InMemoryStorage) TopK(k int) []*proxy.Proxy {
	proxies := make([]*proxy.Proxy, 0)
	s.rbt.Descend(s.rbt.Max(), func(i rbtree.Item) bool {
		proxies = append(proxies, i.(*compareableProxy).pxy)
		return len(proxies) < k
	})
	return proxies
}

func (s *InMemoryStorage) Len() uint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.rbt.Len()
}
