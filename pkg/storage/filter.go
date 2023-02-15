// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"github.com/leosocy/proksi/pkg/proxy"
)

// Filter is used to filter a proxy during the selection process
type Filter func([]*proxy.Proxy) []*proxy.Proxy

// FilterScore is a score based Select Filter which will
// only return proxies which score >= threshold
func FilterScore(threshold int8) Filter {
	return func(old []*proxy.Proxy) []*proxy.Proxy {
		var proxies []*proxy.Proxy
		for _, pxy := range old {
			if pxy.Score >= threshold {
				proxies = append(proxies, pxy)
			}
		}
		return proxies
	}
}
