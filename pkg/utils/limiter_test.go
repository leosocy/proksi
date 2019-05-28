// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	assert := assert.New(t)
	limiter := RateLimiter{Delay: 100 * time.Millisecond, Parallelism: 2}
	limiter.Init()
	var wg sync.WaitGroup
	var costNs int64
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			start := time.Now()
			limiter.Enter()
			defer limiter.Exit()
			time.Sleep(100 * time.Millisecond)
			costNs += time.Now().Sub(start).Nanoseconds()
		}()
	}
	wg.Wait()
	assert.True(costNs > (100*2*2+100-1)*time.Millisecond.Nanoseconds())
}
