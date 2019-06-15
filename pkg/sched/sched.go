// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package sched

import (
	"time"

	"github.com/Leosocy/IntelliProxy/pkg/checker"
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
	"github.com/Leosocy/IntelliProxy/pkg/spider"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
	"github.com/Leosocy/IntelliProxy/pkg/utils"
	"github.com/Sirupsen/logrus"
)

// Scheduler responsible for scheduling cooperation between Spider,Checker and Backend.
type Scheduler struct {
	spiders          []*spider.Spider
	cachedChan       proxy.CachedChan
	scoreChecker     checker.Scorer
	reqHeadersGetter utils.RequestHeadersGetter
	geoInfoFetcher   proxy.GeoInfoFetcher
	backend          storage.Backend
	logger           *logrus.Logger
}

// NewScheduler returns a new scheduler instance with default configuration.
func NewScheduler() *Scheduler {
	sc := &Scheduler{
		spiders:          spider.BuildAndInitAll(),
		cachedChan:       proxy.NewBloomCachedChan(),
		scoreChecker:     checker.NewBatchHTTPSScorer(checker.HostsOfBatchHTTPSScorer),
		reqHeadersGetter: utils.HTTPBinUtil{Timeout: 5 * time.Second},
		geoInfoFetcher:   proxy.NewGeoInfoFetcher(proxy.NameOfIPAPIFetcher),
		backend:          storage.NewInMemoryBackend(),
		logger:           logrus.New(),
	}
	sc.logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	return sc
}

func (sc *Scheduler) GetBackend() storage.Backend {
	return sc.backend
}

// Start open the background crawling, detection, inspection tasks,
// and receive the agent and process.
func (sc *Scheduler) Start() {
	// TODO: threshold从配置中加载
	go sc.bgCrawling(100)
	go sc.bgDetections(15 * time.Minute)
	go sc.bgInspection(30 * time.Minute)
	sc.loopRecv()
}

func (sc *Scheduler) loopRecv() {
	recvCh := sc.cachedChan.Recv()
	for {
		select {
		case pxy := <-recvCh:
			go sc.inspectProxy(pxy)
		}
	}
}

func (sc *Scheduler) inspectProxy(pxy *proxy.Proxy) {
	score := sc.scoreChecker.Score(pxy)
	entry := sc.logger.WithFields(logrus.Fields{
		"url":   pxy.URL(),
		"score": score,
	})
	if score > 0 {
		if inserted, err := sc.backend.InsertOrUpdate(pxy); err == nil {
			action := "Updated"
			if inserted {
				action = "Inserted"
			}
			entry.Infof("%s proxy to backend", action)
		}
	} else {
		if err := sc.backend.Delete(pxy.IP); err == nil {
			entry.Info("Deleted proxy from backend")
		}
	}
}

func (sc *Scheduler) completeProxy(pxy *proxy.Proxy) {
	entry := sc.logger.WithFields(logrus.Fields{
		"url": pxy.URL(),
	})
	if pxy.Anon == proxy.Unknown {
		if err := pxy.DetectAnonymity(sc.reqHeadersGetter); err != nil {
			entry.Warnf("Failed to detect anonymity, %v", err)
		} else {
			if err := sc.backend.Update(pxy); err == nil {
				entry.Info("Updated anonymity")
			}
		}
	}
	if pxy.GeoInfo == nil {
		if err := pxy.DetectGeoInfo(sc.geoInfoFetcher); err != nil {
			entry.Warnf("Failed to detect geography information, %v", err)
		} else {
			if err := sc.backend.Update(pxy); err == nil {
				entry.Infof("Updated geography information")
			}
		}
	}
}

func (sc *Scheduler) bgDetections(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	iterDetections := func() {
		sc.logger.Info("Start iterating the proxies and detecting anonymity/geo info")
		sc.backend.Iter(func(pxy *proxy.Proxy) bool {
			go sc.completeProxy(pxy)
			return true
		})
		sc.logger.Info("Finish iterating the proxies and detecting anonymity/geo info")
	}
	for {
		select {
		case <-ticker.C:
			iterDetections()
		}
	}
}

func (sc *Scheduler) bgInspection(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	iterInspection := func() {
		sc.backend.Iter(func(pxy *proxy.Proxy) bool {
			go sc.inspectProxy(pxy)
			return true
		})
	}
	for {
		select {
		case <-ticker.C:
			iterInspection()
		}
	}
}

// bgCrawling when the number of proxies in backend is less than threshold, start crawling.
func (sc *Scheduler) bgCrawling(threshold uint) {
	for _, s := range sc.spiders {
		go s.Start(sc.cachedChan)
	}
	// TODO: ProxyCountWatcher 监测当代理个数不足阈值时调度spider开启一次爬取
	for {
		for _, s := range sc.spiders {
			s.TryCrawl()
		}
		time.Sleep(20 * time.Minute)
	}
}
