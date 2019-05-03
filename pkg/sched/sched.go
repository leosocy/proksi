// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package sched

import (
	"fmt"
	"sync"

	"github.com/Leosocy/gipp/pkg/utils"

	"github.com/Leosocy/gipp/pkg/checker"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

type Scheduler struct {
	spiders      []*spider.Spider
	cachedChan   proxy.CachedChan
	scoreChecker checker.Scorer
}

// NewScheduler returns a new scheduler instance with default configuration.
func NewScheduler() *Scheduler {
	return &Scheduler{
		spiders:      spider.BuildAndInitAll(),
		cachedChan:   proxy.NewBloomCachedChan(),
		scoreChecker: checker.NewBatchHTTPSScorer(checker.HostsOfBatchHTTPSScorer),
	}
}

// Start starts one goroutine for each spider
// and crawls the proxy to the specified cached channel.
// Receives the proxy in the channel, use the checker
// to score it in a round, and then store it in the specified storage.
func (sc *Scheduler) Start() {
	for _, s := range sc.spiders {
		go s.CrawlTo(sc.cachedChan)
	}
	sc.loopRecv()
}

func (sc *Scheduler) loopRecv() {
	recvCh := sc.cachedChan.Recv()
	for {
		select {
		case pxy := <-recvCh:
			go sc.handleProxy(pxy)
		}
	}
}

func (sc *Scheduler) handleProxy(pxy *proxy.Proxy) {
	score := sc.scoreChecker.Score(pxy)
	if score > 0 {
		sc.doDetections(pxy)
		sc.doSave(pxy)
	}
}

func (sc *Scheduler) doDetections(pxy *proxy.Proxy) {
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		pxy.DetectAnonymity(utils.HTTPBinUtil{})
	}()
	go func() {
		defer wg.Done()
		pxy.DetectGeoInfo(proxy.NewGeoInfoFetcher(proxy.NameOfIPAPIFetcher))
	}()
	go func() {
		defer wg.Done()
		pxy.DetectLatency()
	}()
	go func() {
		defer wg.Done()
		pxy.DetectSpeed()
	}()
	wg.Wait()
}

func (sc *Scheduler) doSave(pxy *proxy.Proxy) {
	// TODO: storage.CreateOrUpdate(pxy)
	fmt.Printf("%+v\n", pxy)
}
