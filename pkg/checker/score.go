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
	Score(pxy *proxy.Proxy)
}

// BatchHTTPSScorer try visiting a batch of HTTPS websites
// and grade the proxy by response time.
type BatchHTTPSScorer struct {
	hosts   []string
	timeout time.Duration
}

// NewBatchHTTPSScorer returns a new scorer, the hosts can't be empty or bigger than maximum score.
// The timeout is calculated by the length of hosts.
// If all host try failed, the proxy score will be reduced to 0
func NewBatchHTTPSScorer(hosts []string) Scorer {
	hl := int8(len(hosts))
	if hl < 2 {
		panic(errors.New("size of hosts can't be smaller than 2"))
	}
	return &BatchHTTPSScorer{
		hosts:   hosts,
		timeout: time.Duration(proxy.MaximumScore/hl*2) * time.Second,
	}
}

// Score try to use proxy visit each host, and modifies
// the corresponding proxy score based on the return value .
func (s *BatchHTTPSScorer) Score(pxy *proxy.Proxy) {
	// since we don't try diff host parallel, so init request here will reduce mem.
	sa := gorequest.New().Proxy(pxy.URL()).Timeout(s.timeout)
	for _, host := range s.hosts {
		rt, _ := s.try(sa, host)
		delta := s.timeout.Seconds()/2 - rt.Seconds()
		pxy.ChangeScore(int8(math.Ceil(delta)))
	}
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
