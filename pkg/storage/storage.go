// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"errors"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

var (
	ErrProxyNil           = errors.New("proxy is nil")
	ErrProxyScoreNegative = errors.New("proxy score <= 0")
)

type QueryCondition struct {
}

type Storage interface {
	Insert(pxy *proxy.Proxy) error
	// Update(pxy *proxy.Proxy) error
	// Delete(pxy *proxy.Proxy) error
	// Query(cond QueryCondition) ([]*proxy.Proxy, error)
	Len() uint
}
