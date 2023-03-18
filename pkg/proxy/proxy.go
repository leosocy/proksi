// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/leosocy/proksi/pkg/geolocation"
	"github.com/leosocy/proksi/pkg/protocol"
	"github.com/leosocy/proksi/pkg/quality"
	"github.com/leosocy/proksi/pkg/traffic"
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

// DetectGeoInfo set the Geolocation field value by calling `NewGeoInfo`
func (p *Proxy) DetectGeoInfo(locator geolocation.Geolocator) (err error) {
	p.Geolocation, err = locator.Locate(context.Background(), p.IP.String())
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

func (p *Proxy) IsValid() bool {
	return p.AddrPort.IsValid() && p.Protocols.IsValid()
}

// URL returns string like `ip:port`
func (p *Proxy) URL() string {
	if len(p.IP) == 0 || p.Port == 0 {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", p.IP.String(), p.Port)
}

func (p *Proxy) String() string {
	return p.AddrPort.String()
}

func (p *Proxy) Equal(to *Proxy) bool {
	return p.AddrPort.Addr().Compare(to.AddrPort.Addr()) == 0 && p.AddrPort.Port() == to.AddrPort.Port()
}

// Builder is a builder pattern code
type Builder struct {
	ip   netip.Addr
	port uint16

	proxy *Proxy
	err   error
}

func NewBuilder() *Builder {
	proxy := &Proxy{
		Protocols: protocol.NothingProtocols,
		Anonymity: AnonymityUnknown,
		CreatedAt: time.Now(),
	}
	b := &Builder{proxy: proxy}
	return b
}

func (b *Builder) IP(ip string) *Builder {
	parsedIP, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		b.err = multierr.Append(b.err, err)
	} else {
		b.ip = parsedIP
	}
	return b
}

func (b *Builder) Port(port string) *Builder {
	parsedPort, err := strconv.ParseUint(strings.TrimSpace(port), 10, 16)
	if err != nil {
		b.err = multierr.Append(b.err, err)
	} else {
		b.port = uint16(parsedPort)
	}
	return b
}

func (b *Builder) AddrPort(s string) *Builder {
	ipp, err := netip.ParseAddrPort(s)
	if err != nil {
		b.err = multierr.Append(b.err, err)
	} else {
		b.ip = ipp.Addr()
		b.port = ipp.Port()
	}
	return b
}

func (b *Builder) Protocols(protocols protocol.Protocols) *Builder {
	b.proxy.Protocols = protocols
	return b
}

func (b *Builder) Anonymity(anonymity Anonymity) *Builder {
	b.proxy.Anonymity = anonymity
	return b
}

func (b *Builder) Quality(quality quality.Quality) *Builder {
	b.proxy.Quality = quality
	return b
}

func (b *Builder) Geolocation(geolocation *geolocation.Geolocation) *Builder {
	b.proxy.Geolocation = geolocation
	return b
}

func (b *Builder) Build() (*Proxy, error) {
	if b.err != nil {
		return nil, b.err
	}

	b.proxy.AddrPort = netip.AddrPortFrom(b.ip, b.port)
	if !b.proxy.AddrPort.IsValid() {
		return nil, errors.New("invalid ip:port " + strconv.Quote(b.proxy.AddrPort.String()))
	}

	return b.proxy, nil
}

func (b *Builder) MustBuild() *Proxy {
	pxy, err := b.Build()
	if err != nil {
		panic(err)
	}
	return pxy
}
