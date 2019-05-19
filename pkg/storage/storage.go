// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"errors"
	"net"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

var (
	ErrProxyInvalid       = errors.New("proxy is nil or score <= 0")
	ErrProxyDuplicated    = errors.New("proxy is already in storage")
	ErrProxyDoesNotExists = errors.New("proxy doesn't exists")
)

type QueryCondition struct {
}

type Storage interface {
	Insert(pxy *proxy.Proxy) error
	Search(ip net.IP) *proxy.Proxy
	Update(pxy *proxy.Proxy) error
	// Query(cond QueryCondition) ([]*proxy.Proxy, error)
	// Delete(pxy *proxy.Proxy) error
	Len() uint
}
