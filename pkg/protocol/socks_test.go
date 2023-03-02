// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSOCKS4Prober_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newSOCKS4Prober()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	protocols, err := prober.Probe(ctx, "220.95.248.56:1080")
	assert.Nil(err)
	assert.Equal(NewProtocols(SOCKS4), protocols)

	protocols, err = prober.Probe(ctx, "142.54.228.193:4145")
	assert.NotNil(err)
	assert.Equal(EmptyProtocols, protocols)
}

func TestSocks5Prober_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newSOCKS5Prober()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	protocols, err := prober.Probe(ctx, "220.95.248.56:1080")
	assert.Nil(err)
	assert.Equal(NewProtocols(SOCKS5), protocols)

	protocols, err = prober.Probe(ctx, "142.54.228.193:4145")
	assert.NotNil(err)
	assert.Equal(EmptyProtocols, protocols)
}
