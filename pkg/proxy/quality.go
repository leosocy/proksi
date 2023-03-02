// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"time"
)

// Uptime represents the amount of time it is operational within a given period.
// This is usually expressed as a percentage, e.g. “Our proxy servers have an uptime of 99.9%”.
type Uptime float64

type Quality struct {
	Latency time.Duration `json:"latency"`
	Uptime  Uptime        `json:"uptime"`
}

type Qualifier interface {
	Qualify(ctx context.Context, proxy *Proxy) (Quality, error)
}
