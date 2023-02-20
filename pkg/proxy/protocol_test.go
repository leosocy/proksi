// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProtocol(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(HTTP, ParseProtocol("HTtp"))
	assert.Equal(HTTP, ParseProtocol(" HTtp"))
	assert.Equal(HTTP, ParseProtocol("HTTP "))
	assert.Equal(HTTPS, ParseProtocol("HTTPS"))
	assert.Equal(SOCKS4, ParseProtocol("SOCKs4 "))
	assert.Equal(SOCKS5, ParseProtocol("SOCKS 5"))
	assert.Equal(ProtocolUnsupported, ParseProtocol("SOCKS 6"))
}

func BenchmarkParseProtocol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseProtocol(" SOCKS 4")
	}
}

func TestProtocols(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Protocols(uint8(HTTP)|uint8(HTTPS)), NewProtocols(HTTP, HTTPS))
	assert.True(NewProtocols(HTTP, HTTPS).Contains(HTTP))
	assert.True(NewProtocols(HTTP, HTTPS).Contains(HTTPS))
	assert.False(NewProtocols(HTTP, HTTPS).Contains(SOCKS4))
}

func BenchmarkNewProtocols(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewProtocols(HTTP, HTTPS)
	}
}

func BenchmarkProtocols_Contains(b *testing.B) {
	protocols := NewProtocols(HTTP, HTTPS)
	for i := 0; i < b.N; i++ {
		protocols.Contains(HTTP)
		protocols.Contains(SOCKS4)
	}
}
