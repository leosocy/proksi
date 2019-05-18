// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
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

type InMemoryStorage struct {
	rbt *rbtree.Rbtree
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		rbt: rbtree.New(),
	}
}

func (s *InMemoryStorage) Insert(pxy *proxy.Proxy) error {
	if pxy == nil || pxy.Score <= 0 {
		return ErrProxyScoreNegative
	}
	s.rbt.Insert(&compareableProxy{pxy: pxy})
	return nil
}

func (s *InMemoryStorage) Len() uint {
	return s.rbt.Len()
}
