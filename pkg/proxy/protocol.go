// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import "strings"

// Protocol represents the protocol types of the proxy.
// Includes HTTP, HTTPS, SOCKS4, and SOCKS5.
type Protocol uint8

const (
	// HTTP proxy can handle HTTP traffic, after receiving it, the proxy connects to the server
	// according to the target domain name or IP address in headers and forwards the client's request to the server,
	// and finally forwards the server's response to the client. All traffic is transmitted in plaintext.
	//
	// The HTTP proxy can also handle HTTPS traffic, and the client sends a CONNECT request to the proxy
	// to indicate the server to connect to. After receive the request, the proxy will connect to the target server
	// and establish a TCP connection. After that, the encrypted data of the client will be forwarded to the server,
	// and the encrypted data of the server will also be forwarded to the client.
	// The proxy does not know the plaintext of the encrypted data.
	HTTP Protocol = 1 << 0

	// HTTPS proxy is similar to an HTTP proxy in that both can handle HTTP and HTTPS traffic.
	// It uses TLS to encrypt data from the client to the proxy,
	// preventing the CONNECT target address from being monitored
	// But currently there are very few HTTPS proxies.
	HTTPS Protocol = 1 << 1

	// SOCKS4 proxy can handle HTTP(s) and TCP traffic.
	SOCKS4 Protocol = 1 << 2

	// SOCKS5 proxy can handle HTTP(s), TCP and UDP traffic. Supports authentication and ipv6.
	// See https://www.rfc-editor.org/rfc/rfc1928 for more details.
	SOCKS5 Protocol = 1 << 3

	// ProtocolUnsupported represents the protocol is not supported.
	ProtocolUnsupported Protocol = 1 << 7
)

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
		return "unknown"
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
		return ProtocolUnsupported
	}
}

// Protocols represents a list of protocols by bitwise.
type Protocols uint8

// Contains returns whether the protocol is included
func (protos Protocols) Contains(proto Protocol) bool {
	return (uint8(protos) & uint8(proto)) != 0
}

func NewProtocols(protos ...Protocol) Protocols {
	var v uint8
	for _, proto := range protos {
		v |= uint8(proto)
	}

	return Protocols(v)
}
