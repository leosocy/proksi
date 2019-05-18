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

func (s *InMemoryStorage) insert(pxy *proxy.Proxy) {
	hasher := fnv.New64()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rbt.Insert(&compareableProxy{pxy: pxy})
	hasher.Write(pxy.IP)
	s.m[hasher.Sum64()] = pxy
}

func (s *InMemoryStorage) Insert(pxy *proxy.Proxy) error {
	if pxy == nil || pxy.Score <= 0 {
		return ErrProxyInvalid
	}
	if sp := s.Search(pxy.IP); sp != nil {
		return ErrProxyDuplicated
	}
	s.insert(pxy)
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

func (s *InMemoryStorage) Len() uint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.rbt.Len()
}
