// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package checker

import "time"

// UsabilityChecker 用于检查代理可用性
type UsabilityChecker interface {
	ProxyUsable(proxyURL string) bool
}

// HTTPSUsabilityChecker 用于检查代理是否可以访问HTTPS网站
type HTTPSUsabilityChecker interface {
	ProxyHTTPSUsable(proxyURL string) bool
}

var (
	// DefaultHosts 用于检测代理是否可访问这些HTTP(s)网站
	DefaultHosts = []string{
		"www.baidu.com",
		"www.liepin.com",
		"lagou.com",
		"zhilian.com",
		"bj.zu.anjuke.com",
		"github.com",
		"blog.csdn.net",
		"movie.douban.com",
		"www.ctrip.com",
		"www.qunar.com",
	}
)

type TimeoutUsabilityChecker struct {
	host    string
	timeout time.Duration
}
