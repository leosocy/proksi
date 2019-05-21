// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"errors"
	"net"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

// Errors occur when using storage.
var (
	ErrProxyInvalid       = errors.New("proxy is nil or score <= 0")
	ErrProxyDuplicated    = errors.New("proxy is already in storage")
	ErrProxyDoesNotExists = errors.New("proxy doesn't exists")
)

// Iterator is the function which will be call for each proxy in storage.
// It will stop whenever the iterator returns false.
type Iterator func(pxy *proxy.Proxy) bool

// Storage is a container for proxies.
type Storage interface {
	Insert(p *proxy.Proxy) error
	Update(newP *proxy.Proxy) error
	InsertOrUpdate(p *proxy.Proxy) error
	Search(ip net.IP) *proxy.Proxy
	Delete(ip net.IP) error
	Len() uint
	// TopK returns the first K proxies order by score descend.
	TopK(k int) []*proxy.Proxy
	Iter(iter Iterator)
	// Query(cond QueryCondition) ([]*proxy.Proxy, error)
}

type QueryCondition struct {
}
