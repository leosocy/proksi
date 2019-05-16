// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package sched

import "time"

// LimitRule contains some limit for scheduler when processing proxy.
type LimitRule struct {
	// Delay is the duration to wait before handle proxy.
	Delay time.Duration
	// Parallelism is the number of the maximum allowed concurrent handling.
	Parallelism int
	waitChan    chan struct{}
}

// Init initializes the private members of LimitRule
func (r *LimitRule) Init() error {
	waitChanSize := 1
	if r.Parallelism > 1 {
		waitChanSize = r.Parallelism
	}
	r.waitChan = make(chan struct{}, waitChanSize)
	return nil
}
