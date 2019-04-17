// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"errors"
	"net"
	"strconv"
	"time"
)

// Anonymity 匿名度, 请求`https://httpbin.org/get?show_env=1`
// 根据ResponseBody中的 `X-Real-IP` 和 `Via`字段判断。
// 另外如果代理支持HTTPS，访问https网站是没有匿名度的概念的，
// 因为此时代理只负责传输数据，并不能解析替换RequestHeaders。
type Anonymity uint8

const (
	// Transparent 透明：服务器知道你使用了代理，并且能查到原始IP
	Transparent Anonymity = 1
	// Anonymous 普通匿名(较为少见)：服务器知道你使用了代理，但是查不到原始IP
	Anonymous Anonymity = 2
	// Elite 高级匿名：服务器不知道你使用了代理
	Elite Anonymity = 3 // 高匿名
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
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	proxyScoreMaximum uint = 100
)

// NewProxy passes in the ip, port,
// calculates the other field values,
// and returns an initialized Proxy object
func NewProxy(ip, port string) (*Proxy, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.New("invalid ip")
	}
	parsedPort, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		IP:        parsedIP,
		Port:      uint32(parsedPort),
		Score:     proxyScoreMaximum,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
