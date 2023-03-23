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

func TestCompositeProber_Probe(t *testing.T) {
	assert := assert.New(t)
	prober := newCombinedProber()
	ctx, _ := context.WithTimeout(context.Background(), 7*time.Second)
	protocol, err := prober.Probe(ctx, "112.54.47.55:9091")
	assert.Nil(err)
	assert.Equal(HTTP, protocol)

	protocol, err = prober.Probe(ctx, "142.54.228.193:6666")
	assert.NotNil(err)
	assert.Equal(Nothing, protocol)
}
