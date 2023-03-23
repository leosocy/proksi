// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"strings"
)

// Protocol represents the type of protocol that a proxy supports.
//
// Additionally, a Proxy's protocol can support different types of traffic.Traffic.
// For example, a Proxy with HTTP protocol can support both traffic.HTTP and traffic.HTTPS traffic,
// while a Proxy with SOCKS4 protocol can support all traffic.TCP traffic.
//
// Includes HTTP, HTTPS, SOCKS4, and SOCKS5.
type Protocol uint8

const (
	// HTTP means the client communicates with the proxy through the HTTP protocol,
	// after receiving traffic.HTTP traffic, the proxy connects to the target server according to the host
	// in headers and forwards the client's request to the server,
	// and finally forwards the server's response to the client.
	// All traffic is transmitted in plaintext.
	//
	// It can also handle https traffic, the client sends a CONNECT request to the proxy
	// to indicate the server to connect to. After receive the request, the proxy will connect to the target server
	// and establish a TCP connection. After that, the encrypted data of the client will be forwarded to the server,
	// and the encrypted data of the server will also be forwarded to the client.
	// The proxy does not know the plaintext of the encrypted data.
	// But some HTTP proxy will return 400 when CONNECT.
	HTTP Protocol = 1 << 0

	// HTTPS means the client communicates with the proxy through the HTTPS protocol,
	// it uses TLS to encrypt data from the client to the proxy,
	// preventing the CONNECT target address from being monitored.
	HTTPS Protocol = 1 << 1

	// SOCKS4 means the client communicates with the proxy through the SOCKS4 protocol,
	// It allows a client to connect to a server and request that the server establish a connection to
	// another host on behalf of the client, it can handle traffic.HTTP(s) and traffic.TCP traffic.
	//
	// SOCKS4 does not support authentication, and it only supports IPv4 addresses.
	// It has largely been replaced by SOCKS5, which adds support for authentication, IPv6 addresses, and other features.
	SOCKS4 Protocol = 1 << 2

	// SOCKS5 means the client communicates with the proxy through the SOCKS5 protocol,
	// proxy can handle traffic.HTTP(s), traffic.TCP and UDP traffic. Supports authentication and ipv6.
	// See https://www.rfc-editor.org/rfc/rfc1928 for more details.
	SOCKS5 Protocol = 1 << 3

	// Nothing means the protocol is unknown.
	Nothing Protocol = 0
)

func (proto Protocol) IsValid() bool {
	return (proto & 0x0f) != 0
}

// String returns a string representation of Protocol
func (proto Protocol) String() string {
	switch proto {
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	case SOCKS4:
		return "socks4"
	case SOCKS5:
		return "socks5"
	default:
		return "nothing"
	}
}

// ParseProtocol parse protocol from string
func ParseProtocol(s string) Protocol {
	s = strings.ToLower(strings.ReplaceAll(s, ` `, ``))
	switch s {
	case "http":
		return HTTP
	case "https":
		return HTTPS
	case "socks4":
		return SOCKS4
	case "socks5":
		return SOCKS5
	default:
		return Nothing
	}
}

// Prober is an interface for detecting which protocol a proxy server supports.
type Prober interface {
	// Probe returns the protocol supported by the proxy.
	// If no protocol are supported, return Nothing and an error represents why not support.
	// The addr parameter should be in the format "host:port", for example "1.2.3.4:1080", "proksi.io:1080"
	Probe(ctx context.Context, addr string) (Protocol, error)
}