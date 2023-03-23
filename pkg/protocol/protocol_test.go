// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

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
	assert.Equal(Nothing, ParseProtocol("SOCKS 6"))
}

func BenchmarkParseProtocol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseProtocol(" SOCKS 4")
	}
}

func FuzzParseProtocol(f *testing.F) {
	testdata := []string{"htTp", "SOCKS4"}
	for _, d := range testdata {
		f.Add(d)
	}

	f.Fuzz(func(t *testing.T, data string) {
		p := ParseProtocol(data)
		doubleP := ParseProtocol(p.String())
		assert.True(t, p == doubleP)
	})
}
