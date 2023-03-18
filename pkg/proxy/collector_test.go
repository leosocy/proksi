// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testCollector struct {
	numOfCalls uint16
	proxies    []*Proxy
	lock       sync.Mutex
}

func (c *testCollector) Collect(ps ...*Proxy) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.numOfCalls += 1
	c.proxies = append(c.proxies, ps...)
}

func (c *testCollector) Close() error {
	return nil
}

func TestBatchedCollector_BatchSize(t *testing.T) {
	assert := assert.New(t)
	c := &testCollector{}
	batched := NewBatchedCollector(c, 10*time.Second, 2)
	batched.Collect(nil)
	assert.Equal(0, int(c.numOfCalls))
	batched.Collect(nil)
	// sleep to wait async accumulate
	time.Sleep(10 * time.Millisecond)
	assert.Equal(1, int(c.numOfCalls))
}

func TestBatchedCollector_WaitTime(t *testing.T) {
	assert := assert.New(t)
	c := &testCollector{}
	batched := NewBatchedCollector(c, 10*time.Millisecond, 2)
	batched.Collect(nil)
	assert.Equal(0, int(c.numOfCalls))
	// sleep to wait async accumulate time tick
	time.Sleep(15 * time.Millisecond)
	assert.Equal(1, int(c.numOfCalls))
}

func TestBatchedCollector_Close(t *testing.T) {
	assert := assert.New(t)
	c := &testCollector{}
	batched := NewBatchedCollector(c, 10*time.Second, 2)
	batched.Collect(nil)
	assert.Equal(0, int(c.numOfCalls))
	batched.Close()
	// no need to sleep, because Close will wait all data recv in chan.
	assert.Equal(1, int(c.numOfCalls))
}
