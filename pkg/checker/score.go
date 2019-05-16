// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package checker

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/parnurzeal/gorequest"
)

// Scorer is the interface used to score a proxy.
type Scorer interface {
	// Score calculates the proxy's score.
	Score(pxy *proxy.Proxy) int8
}

// BatchHTTPSScorer try visiting a batch of HTTPS websites
// and grade the proxy by response time.
type BatchHTTPSScorer struct {
	hosts   []string
	timeout time.Duration
}

// NewBatchHTTPSScorer returns a new scorer,
// the hosts can't be empty or bigger than maximum score.
// The timeout is calculated by the length of hosts.
// If response time smaller than timeout/2,
// score increments by (timeout/2 - RT)
// else score decrements by (RT - timeout/2).
// So, if all host try failed, the proxy score will be reduced to 0
func NewBatchHTTPSScorer(hosts []string) Scorer {
	if len(hosts) < 2 {
		panic(errors.New("length of hosts must be bigger than 2"))
	}
	// Ceil to make sure that the score is reduced to 0 when all try fails.
	avg := math.Ceil(float64(proxy.MaximumScore) / float64(len(hosts)))
	return &BatchHTTPSScorer{
		hosts:   hosts,
		timeout: time.Duration(avg*2) * time.Second,
	}
}

// Score try to use proxy visit each host, and modifies
// the corresponding proxy score based on the return value .
func (s *BatchHTTPSScorer) Score(pxy *proxy.Proxy) int8 {
	// since we don't try diff host parallel, so init request here to reduce mem cost.
	sa := gorequest.New().Proxy(pxy.URL()).Timeout(s.timeout)
	for _, host := range s.hosts {
		rt, _ := s.try(sa, host)
		delta := (s.timeout/2 - rt).Seconds()
		pxy.AddScore(int8(math.Floor(delta)))
	}
	return pxy.Score
}

// do requests to host with proxy and timeout, then calculate the response time.
func (s *BatchHTTPSScorer) try(sa *gorequest.SuperAgent, host string) (rt time.Duration, err error) {
	start := time.Now()
	resp, _, errs := sa.Get(host).EndBytes()
	if resp == nil || resp.StatusCode != http.StatusOK || errs != nil {
		err = fmt.Errorf("try to request host %s failed", host)
		rt = s.timeout
	} else {
		rt = time.Since(start)
	}
	return
}
