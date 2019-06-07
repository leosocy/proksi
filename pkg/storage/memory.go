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

func (s *InMemoryStorage) insert(p *proxy.Proxy) error {
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

func (s *InMemoryStorage) Insert(p *proxy.Proxy) error {
	if p == nil || p.Score <= 0 {
		return ErrProxyInvalid
	}
	if sp := s.Search(p.IP); sp != nil {
		return ErrProxyDuplicated
	}
	return s.insert(p)
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

func (s *InMemoryStorage) delete(p *proxy.Proxy) error {
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

func (s *InMemoryStorage) Delete(ip net.IP) error {
	var sp *proxy.Proxy
	if sp = s.Search(ip); sp == nil {
		return ErrProxyDoesNotExists
	}
	return s.delete(sp)
}

// Update implements storage.Update.
func (s *InMemoryStorage) Update(newP *proxy.Proxy) error {
	if err := s.Delete(newP.IP); err != nil {
		return err
	}
	return s.Insert(newP)
}

func (s *InMemoryStorage) InsertOrUpdate(p *proxy.Proxy) (bool, error) {
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

func (s *InMemoryStorage) Len() uint {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.rbt.Len()
}

func (s *InMemoryStorage) TopK(k int) []*proxy.Proxy {
	proxies := make([]*proxy.Proxy, 0)
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.rbt.Descend(s.rbt.Max(), func(item rbtree.Item) bool {
		proxies = append(proxies, item.(*comparableProxy).pxy)
		return k == 0 || len(proxies) < k
	})
	return proxies
}

func (s *InMemoryStorage) Iter(iter Iterator) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.rbt.Ascend(s.rbt.Min(), func(item rbtree.Item) bool {
		pxy := item.(*comparableProxy).pxy
		return iter(pxy)
	})
}
