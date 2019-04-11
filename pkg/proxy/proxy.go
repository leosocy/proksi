// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net"
	"time"
)

// AnonymityLevel 匿名度, https://httpbin.org/get\?show_env\=1
type AnonymityLevel uint8

// ProtocolType 代理类型
type ProtocolType uint8

const (
	// Transparent 透明：服务器知道你使用了代理，并且能查到原始IP
	Transparent AnonymityLevel = 1
	// Anonymous 普通匿名：服务器知道你使用了代理，但是查不到原始IP
	Anonymous AnonymityLevel = 2
	// HighAnonymous 高级匿名：服务器不知道你使用了代理
	HighAnonymous AnonymityLevel = 3 // 高匿名
)

const (
	// HTTP 不能访问只支持https的网站，并且会有数据被代理服务器监听的风险
	HTTP ProtocolType = 1
	// HTTPS 可以访问https网站，并且代理服务器无法截获数据
	HTTPS ProtocolType = 2
)

// Proxy IP Proxy data model.
type Proxy struct {
	IP        net.IP         `json:"ip"`
	Port      uint32         `json:"port"`
	Loc       string         `json:"location"`
	AL        AnonymityLevel `json:"anonymity_level"`
	PT        ProtocolType   `json:"protocol_type"`
	Speed     uint32         `json:"speed"` // unit: ms
	Score     uint           `json:"score"` // full is 100
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
