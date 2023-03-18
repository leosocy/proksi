// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package traffic

import (
	"context"
	"strings"
)

// Traffic describes which types of traffic that the proxy can handle.
// For example, an protocol.HTTP proxy can handle plaintext HTTP traffic, while an HTTPS proxy can handle encrypted HTTPS traffic.
// Includes HTTP, HTTPS, TCP, and SOCKS5.
type Traffic uint8

const (
	// HTTP means the proxy can handle http traffic, after receiving it, the proxy connects to the server
	// according to the target domain name or IP address in headers and forwards the client's request to the server,
	// and finally forwards the server's response to the client. All traffic is transmitted in plaintext.
	//
	// Although strictly speaking, HTTP proxy can also handle https traffic,
	// but considering that there are actually few HTTPS proxies,
	// in proksi, HTTP means that the proxy can handle http traffic.
	HTTP Traffic = 1 << 1

	// HTTPS means proxy can handle https traffic, and the client sends a CONNECT request to the proxy
	// to indicate the server to connect to. After receive the request, the proxy will connect to the target server
	// and establish a TCP connection. After that, the encrypted data of the client will be forwarded to the server,
	// and the encrypted data of the server will also be forwarded to the client.
	// The proxy does not know the plaintext of the encrypted data.
	//
	// Although strictly speaking, HTTPS proxy means it using TLS to encrypt data from the client to the proxy,
	// preventing the CONNECT target address from being monitored,
	// but considering that there are actually few HTTPS proxies,
	// in proksi, HTTPS means that the proxy can handle https traffic.
	HTTPS Traffic = 1 << 2

	// TCP proxy can handle HTTP(s) and TCP traffic.
	TCP Traffic = 1 << 3

	UDP Traffic = 1 << 4
)

// String returns a string representation of Traffic
func (proto Traffic) String() string {
	switch proto {
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	default:
		return "unknown"
	}
}

// ParseTraffic parse protocol from string
func ParseTraffic(s string) Traffic {
	s = strings.ToLower(strings.ReplaceAll(s, ` `, ``))
	switch s {
	case "http":
		return HTTP
	case "https":
		return HTTPS
	default:
		return HTTPS
	}
}

// Traffics represents a list of protocols by bitwise.
type Traffics uint8

const (
	EmptyTraffics Traffics = 0
)

// Supports returns whether the protocol is supported
func (protos Traffics) Supports(proto Traffic) bool {
	return (uint8(protos) & uint8(proto)) != 0
}

func (protos Traffics) Combine(other Traffics) Traffics {
	return protos | other
}

func (protos Traffics) String() string {
	names := make([]string, 0, 2)
	for i := 0; i < 8; i++ {
		if (1<<i)&protos != 0 {
			names = append(names, Traffic(1<<i).String())
		}
	}

	return strings.Join(names, ",")
}

// NewTraffics receives one or more Traffic and
// returns a new Traffics with all the Traffic combined.
func NewTraffics(protos ...Traffic) Traffics {
	var v uint8
	for _, proto := range protos {
		v |= uint8(proto)
	}

	return Traffics(v)
}

// TrafficProber is an interface for detecting which protocols a proxy server supports.
type TrafficProber interface {
	// Probe returns the protocols supported by the proxy.
	// The addr parameter should be in the format "host:port", for example "1.2.3.4:1080", "proksi.io:1080"
	Probe(ctx context.Context, addr string) (Traffics, error)
}
