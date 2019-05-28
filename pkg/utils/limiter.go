// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import "time"

// Limiter
type Limiter interface {
	Init()
	Enter()
	Exit()
}

// RateLimiter contains some rate limit to control
// the rate of some operations, such as network request frequency, etc.
type RateLimiter struct {
	// Delay is the duration to wait before do next operation.
	Delay time.Duration
	// Parallelism is the number of the maximum allowed concurrent operate.
	Parallelism int
	waitChan    chan struct{}
}

// Init initializes the private members of RateLimiter.
func (l *RateLimiter) Init() {
	waitChanSize := 1
	if l.Parallelism > 1 {
		waitChanSize = l.Parallelism
	}
	l.waitChan = make(chan struct{}, waitChanSize)
}

// Enter implements Limiter.Enter.
func (l *RateLimiter) Enter() {
	l.waitChan <- struct{}{}
}

// Exit implements Limiter.Exit.
func (l *RateLimiter) Exit() {
	time.Sleep(l.Delay)
	<-l.waitChan
}
