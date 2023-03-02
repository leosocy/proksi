// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import "strings"

// Anonymity represents the anonymity level of the proxy.
//
// If the proxy supports HTTPS protocol, the anonymity is meaningless,
// because at this time the proxy only transmits encrypted data and cannot parse the request header,
// so the server cannot detect whether the request uses a proxy.
type Anonymity uint8

const (
	// Transparent means the server knows that you use a proxy and can find the original IP
	Transparent Anonymity = 10

	// Anonymous means the server knows that you use a proxy, but the original IP cannot be found
	Anonymous Anonymity = 20

	// Elite means the server doesn't know you're using a proxy
	Elite Anonymity = 30

	// AnonymityUnknown means the anonymity is unknown
	AnonymityUnknown Anonymity = 255
)

// String returns a string representation of Anonymity.
func (anno Anonymity) String() string {
	switch anno {
	case Elite:
		return "elite"
	case Anonymous:
		return "anonymous"
	case Transparent:
		return "transparent"
	default:
		return "unknown"
	}
}

func ParseAnonymity(s string) Anonymity {
	s = strings.ToLower(strings.ReplaceAll(s, ` `, ""))
	switch s {
	case "elite":
		return Elite
	case "anonymous":
		return Anonymous
	case "transparent":
		return Transparent
	default:
		return AnonymityUnknown
	}
}
