// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSOCKS4Prober_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newSOCKS4Prober()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	protocol, err := prober.Probe(ctx, "220.95.248.56:1080")
	assert.Nil(err)
	assert.Equal(SOCKS4, protocol)

	protocol, err = prober.Probe(ctx, "142.54.228.193:4145")
	assert.NotNil(err)
	assert.Equal(Nothing, protocol)
}

func TestSocks5Prober_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newSOCKS5Prober()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	protocol, err := prober.Probe(ctx, "220.95.248.56:1080")
	assert.Nil(err)
	assert.Equal(SOCKS5, protocol)

	protocol, err = prober.Probe(ctx, "142.54.228.193:4145")
	assert.NotNil(err)
	assert.Equal(Nothing, protocol)
}
