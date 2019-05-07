// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package sched

import (
	"fmt"
	"sync"
	"time"

	"github.com/Leosocy/gipp/pkg/utils"

	"github.com/Leosocy/gipp/pkg/checker"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

// Scheduler responsible for scheduling cooperation between Spider,Checker and Storage.
type Scheduler struct {
	spiders          []*spider.Spider
	cachedChan       proxy.CachedChan
	scoreChecker     checker.Scorer
	reqHeadersGetter utils.RequestHeadersGetter
	geoInfoFetcher   proxy.GeoInfoFetcher
	limiter          *LimitRule
}

// NewScheduler returns a new scheduler instance with default configuration.
func NewScheduler() *Scheduler {
	return &Scheduler{
		spiders:          spider.BuildAndInitAll(),
		cachedChan:       proxy.NewBloomCachedChan(),
		scoreChecker:     checker.NewBatchHTTPSScorer(checker.HostsOfBatchHTTPSScorer),
		reqHeadersGetter: utils.HTTPBinUtil{},
		geoInfoFetcher:   proxy.NewGeoInfoFetcher(proxy.NameOfIPAPIFetcher),
	}
}

// RateLimit set a new LimitRule to the scheduler.
func (sc *Scheduler) RateLimit(r *LimitRule) error {
	sc.limiter = r
	return r.Init()
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
			go sc.procProxy(pxy)
		}
	}
}

func (sc *Scheduler) procProxy(pxy *proxy.Proxy) {
	if sc.limiter != nil {
		sc.limiter.waitChan <- struct{}{}
		defer func() {
			time.Sleep(sc.limiter.Delay)
			<-sc.limiter.waitChan
		}()
	}
	score := sc.scoreChecker.Score(pxy)
	if score > 0 {
		sc.doDetect(pxy)
		sc.doSave(pxy)
	}
}

// TODO: scheduler只负责质量检查，对于合格的proxy，写入storage后，由一个后台线程定期从中取出记录
// 如果anonymity/GeoInfo 为空则补全，另外会无条件detect speed/latency。
func (sc *Scheduler) doDetect(pxy *proxy.Proxy) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		pxy.DetectAnonymity(sc.reqHeadersGetter)
	}()
	// go func() {
	// 	defer wg.Done()
	// 	pxy.DetectGeoInfo(sc.geoInfoFetcher)
	// }()
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
