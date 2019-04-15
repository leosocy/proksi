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

// Type 匿名度, https://httpbin.org/get\?show_env\=1
type Type uint8

// Protocol 代理类型
type Protocol uint8

const (
	// Transparent 透明：服务器知道你使用了代理，并且能查到原始IP
	Transparent Type = 1
	// Anonymous 普通匿名：服务器知道你使用了代理，但是查不到原始IP
	Anonymous Type = 2
	// HighAnonymous 高级匿名：服务器不知道你使用了代理
	HighAnonymous Type = 3 // 高匿名
)

const (
	// HTTP 不能访问只支持https的网站，并且会有数据被代理服务器监听的风险
	HTTP Protocol = 1
	// HTTPS 可以访问https网站，并且代理服务器无法截获数据
	HTTPS Protocol = 2
	// ALL 支持 HTTP 和 HTTPS
	ALL Protocol = 3
)

// Proxy IP Proxy data model.
type Proxy struct {
	IP        net.IP    `json:"ip"`
	Port      uint32    `json:"port"`
	GeoInfo   *GeoInfo  `json:"geo_info"`
	Type      Type      `json:"type"`
	Proto     Protocol  `json:"protocol"`
	Speed     uint32    `json:"speed"` // unit: ms
	Score     uint      `json:"score"` // full is 100
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewProxy passes in the ip, port, and location strings,
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
	info, _ := NewGeoInfo(ip)
	return &Proxy{
		IP:        parsedIP,
		Port:      uint32(parsedPort),
		GeoInfo:   info,
		Score:     100, // 初始值设置为满分，加入Pool后会由Inspector进行打分
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
