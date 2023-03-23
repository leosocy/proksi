// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package backend

import (
	"errors"
	"net"

	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/leosocy/proksi/pkg/storage"
)

// Errors occur when using backend.
var (
	ErrProxyInvalid       = errors.New("proxy is nil or score <= 0")
	ErrProxyDuplicated    = errors.New("proxy is already in backend")
	ErrProxyDoesNotExists = errors.New("proxy doesn't exists")
	ErrProxyNoneAvailable = errors.New("proxy none available")
)

// Iterator is the function which will be call for each proxy in backend.
// It will stop when the iterator returns false.
type Iterator func(pxy *proxy.Proxy) bool

// Backend is an interface that store and manipulate proxies
type Backend interface {
	Insert(p *proxy.Proxy) error
	Update(newP *proxy.Proxy) error
	InsertOrUpdate(p *proxy.Proxy) (inserted bool, err error)
	Delete(p *proxy.Proxy) error
	Search(ip net.IP) *proxy.Proxy
	// Select returns proxies after filter with options
	Select(opts ...storage.SelectOption) ([]*proxy.Proxy, error)
	Len() uint
	// TopK returns the first K proxies order by score descend.
	// If k is equal to 0, return all proxies in the backend
	TopK(k int) []*proxy.Proxy
	Iter(iter Iterator)
}
