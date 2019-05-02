// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package checker

// UsabilityChecker 用于检查代理可用性
type UsabilityChecker interface {
	ProxyUsable(proxyURL string) bool
}

// HTTPSUsabilityChecker 用于检查代理是否可以访问HTTPS网站
type HTTPSUsabilityChecker interface {
	ProxyHTTPSUsable(proxyURL string) bool
}
