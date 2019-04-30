// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloomCachedChan(t *testing.T) {
	assert := assert.New(t)
	c := NewBloomCachedChan()
	c.Send("1.2.3.4", "80")
	assert.Equal(1, len(c.Recv()))
	c.Send("5.6.7.8", "80")
	assert.Equal(2, len(c.Recv()))
	// filtered by bloom
	c.Send("5.6.7.8", "80")
	assert.Equal(2, len(c.Recv()))
}

func BenchmarkBloomCachedChan(b *testing.B) {
	c := NewBloomCachedChan()
	for i := 0; i < b.N; i++ {
		c.Send("1.2.3.4", "80")
	}
}
