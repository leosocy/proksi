// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/leosocy/proksi/pkg/geolocation"
	"github.com/leosocy/proksi/pkg/protocol"
	"github.com/leosocy/proksi/pkg/quality"
	"github.com/leosocy/proksi/pkg/traffic"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/leosocy/proksi/pkg/utils"
)

// MaximumScore 代理最大得分
const MaximumScore int8 = 100

// Proxy describes a domain object.
type Proxy struct {
	AddrPort    netip.AddrPort
	Protocols   protocol.Protocols
	Traffic     traffic.Traffics
	Anonymity   Anonymity
	Quality     quality.Quality
	Geolocation *geolocation.Geolocation
	CreatedAt   time.Time
	CheckedAt   time.Time

	// DEPRECATED below
	IP    net.IP
	Port  uint32
	Score int8

	lock sync.RWMutex
}

// NewProxy passes in the ip, port, calculates the other field values,
// and returns an initialized Proxy object
func NewProxy(ip, port string) (*Proxy, error) {
	if ip == "" || port == "" {
		return nil, errors.New("empty ip or port")
	}
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
		Score:     MaximumScore,
		CreatedAt: time.Now(),
		CheckedAt: time.Now(),
	}, nil
}

// DetectGeoInfo set the Geolocation field value by calling `NewGeoInfo`
func (p *Proxy) DetectGeoInfo(locator geolocation.Geolocator) (err error) {
	p.Geolocation, err = locator.Locate(context.Background(), p.IP.String())
	return
}

// DetectAnonymity use a `utils.RequestHeadersGetter` to get a http request headers,
// and then use the following logic to determine the anonymity
//
// If `X-Real-Ip` is equal to the public ip, the anonymity is `Transparent`.
// If `X-Real-Ip` is not equal to the public ip,
// and `Via` field exists, the anonymity is `Anonymous`.
// Otherwise, the anonymity is `Elite`.
func (p *Proxy) DetectAnonymity(g utils.RequestHeadersGetter) (err error) {
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
	if headersUsingProxy, err = g.GetRequestHeadersUsingProxy(p.URL()); err != nil {
		return
	}
	if publicIPUsingProxy, err = headersUsingProxy.ParsePublicIP(); err != nil {
		return
	}
	if publicIP.Equal(publicIPUsingProxy) {
		p.Anonymity = Transparent
	} else {
		if headersUsingProxy.Via != "" {
			p.Anonymity = Anonymous
		} else {
			p.Anonymity = Elite
		}
	}
	return
}

// DetectLatency TODO: detect proxy lentency by request one website N times,
// and calculate average response time.
func (p *Proxy) DetectLatency() {
}

// DetectSpeed TODO: detect proxy speed by download a large file,
// and calculate speed `kb_of_file_size / download_cost_time = n kb/s`
func (p *Proxy) DetectSpeed() {
}

// AddScore adds delta to proxy's score.
func (p *Proxy) AddScore(delta int8) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if delta > 0 {
		if p.Score > MaximumScore-delta {
			p.Score = MaximumScore
			return
		}
	} else {
		if p.Score < -delta {
			p.Score = 0
			return
		}
	}
	p.Score += delta
	p.CheckedAt = time.Now()
}

// URL returns string like `ip:port`
func (p *Proxy) URL() string {
	if len(p.IP) == 0 || p.Port == 0 {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", p.IP.String(), p.Port)
}

func (p *Proxy) String() string {
	return p.IP.String()
}

func (p *Proxy) Equal(to *Proxy) bool {
	return p.IP.Equal(to.IP)
}

// Proxy builder pattern code
type ProxyBuilder struct {
	proxy *Proxy
}

func NewProxyBuilder() *ProxyBuilder {
	proxy := &Proxy{}
	b := &ProxyBuilder{proxy: proxy}
	return b
}

func (b *ProxyBuilder) AddrPort(addrPort netip.AddrPort) *ProxyBuilder {
	b.proxy.AddrPort = addrPort
	return b
}

func (b *ProxyBuilder) Protocols(protocols protocol.Protocols) *ProxyBuilder {
	b.proxy.Protocols = protocols
	return b
}

func (b *ProxyBuilder) Anonymity(anonymity Anonymity) *ProxyBuilder {
	b.proxy.Anonymity = anonymity
	return b
}

func (b *ProxyBuilder) Quality(quality quality.Quality) *ProxyBuilder {
	b.proxy.Quality = quality
	return b
}

func (b *ProxyBuilder) Geolocation(geolocation *geolocation.Geolocation) *ProxyBuilder {
	b.proxy.Geolocation = geolocation
	return b
}

func (b *ProxyBuilder) Port(port uint32) *ProxyBuilder {
	b.proxy.Port = port
	return b
}

func (b *ProxyBuilder) Score(score int8) *ProxyBuilder {
	b.proxy.Score = score
	return b
}

func (b *ProxyBuilder) Build() (*Proxy, error) {
	return b.proxy, nil
}
