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

func TestCombinedProber_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newCombinedProber()
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	protocols, err := prober.Probe(ctx, "220.95.248.56:1080")
	assert.Nil(err)
	assert.Equal(NewProtocols(SOCKS5, SOCKS4), protocols)

	protocols, err = prober.Probe(ctx, "142.54.228.193:6666")
	assert.Nil(err)
	assert.Equal(NothingProtocols, protocols)
}
