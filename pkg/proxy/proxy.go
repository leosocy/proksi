// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Leosocy/gipp/pkg/checker"

	"github.com/Leosocy/gipp/pkg/utils"
)

// Anonymity 匿名度, 请求`https://httpbin.org/get?show_env=1`
// 根据ResponseBody中的 `X-Real-Ip` 和 `Via`字段判断。
// 另外如果代理支持HTTPS，访问https网站是没有匿名度的概念的，
// 因为此时代理只负责传输数据，并不能解析替换RequestHeaders。
type Anonymity uint8

const (
	// Unknown 探测不到匿名度
	Unknown Anonymity = 0
	// Transparent 透明：服务器知道你使用了代理，并且能查到原始IP
	Transparent Anonymity = 1
	// Anonymous 普通匿名(较为少见)：服务器知道你使用了代理，但是查不到原始IP
	Anonymous Anonymity = 2
	// Elite 高级匿名：服务器不知道你使用了代理
	Elite             Anonymity = 3 // 高匿名
	proxyScoreMaximum uint      = 100
)

// Proxy IP Proxy data model.
type Proxy struct {
	IP        net.IP    `json:"ip"`
	Port      uint32    `json:"port"`
	GeoInfo   *GeoInfo  `json:"geo_info"`
	Anon      Anonymity `json:"anonymity"`
	HTTPS     bool      `json:"support_https"` // 是否支持访问https网站
	Latency   uint32    `json:"latency"`       // unit: ms
	Speed     uint32    `json:"speed"`         // unit: kb/s
	Score     uint      `json:"score"`         // full is 100
	CreatedAt time.Time `json:"created_at"`
	CheckedAt time.Time `json:"checked_at"`
}

// NewProxy passes in the ip, port,
// calculates the other field values,
// and returns an initialized Proxy object
func NewProxy(ip, port string) (*Proxy, error) {
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	if parsedIP == nil {
		return nil, errors.New("invalid ip")
	}
	parsedPort, err := strconv.ParseUint(strings.TrimSpace(port), 10, 32)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		IP:        parsedIP,
		Port:      uint32(parsedPort),
		Score:     proxyScoreMaximum,
		CreatedAt: time.Now(),
		CheckedAt: time.Now(),
	}, nil
}

// DetectGeoInfo set the GeoInfo field value by calling `NewGeoInfo`
func (pxy *Proxy) DetectGeoInfo(f GeoInfoFetcher) (err error) {
	pxy.GeoInfo, err = f.Do(pxy.IP.String())
	return
}

// DetectAnonymity use a `utils.RequestHeadersGetter` to get a http request headers,
// and then use the following logic to determine the anonymity
//
// If `X-Real-Ip` is equal to the public ip, the anonymity is `Transparent`.
// If `X-Real-Ip` is not equal to the public ip,
// and `Via` field exists, the anonymity is `Anonymous`.
// Otherwise, the anonymity is `Elite`.
func (pxy *Proxy) DetectAnonymity(g utils.RequestHeadersGetter) (err error) {
	var (
		headers, headersUsingProxy   utils.HTTPRequestHeaders
		publicIP, publicIPUsingProxy net.IP
	)
	if headers, err = g.GetRequestHeaders(); err != nil {
		return
	}
	if publicIP, err = headers.ParsePublicIP(); err != nil {
		return
	}
	if headersUsingProxy, err = g.GetRequestHeadersUsingProxy(pxy.URL()); err != nil {
		return
	}
	if publicIPUsingProxy, err = headersUsingProxy.ParsePublicIP(); err != nil {
		return
	}
	if publicIP.Equal(publicIPUsingProxy) {
		pxy.Anon = Transparent
	} else {
		if headersUsingProxy.Via != "" {
			pxy.Anon = Anonymous
		} else {
			pxy.Anon = Elite
		}
	}
	return
}

// DetectHTTPSSupported use a `checker.HTTPSUsabilityChecker` to
// check whether the proxy can access HTTPS sites.
func (pxy *Proxy) DetectHTTPSSupported(c checker.HTTPSUsabilityChecker) {
	pxy.HTTPS = c.ProxyHTTPSUsable(pxy.URL())
}

func (pxy *Proxy) DetectLatencyAndSpeed() {

}

// URL returns string like `ip:port`
func (pxy *Proxy) URL() string {
	return fmt.Sprintf("http://%s:%d", pxy.IP.String(), pxy.Port)
}
