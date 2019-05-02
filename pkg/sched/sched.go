// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package sched

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/checker"

	"github.com/Leosocy/gipp/pkg/utils"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

type Scheduler struct {
	spiders      []*spider.Spider
	cachedChan   *proxy.CachedChan
	scoreChecker *checker.Scorer
}

// Start starts a goroutine for each spider and starts
// crawling the proxy to the specified cached channel.
// Receives the proxy in the channel, use the checker
// to score it in a round, and then store it in the specified storage.
func Start() {
	totalCount := 0
	ch := proxy.NewBloomCachedChan()
	spiders := spider.BuildAndInitAll()
	for _, s := range spiders {
		go s.CrawlTo(ch)
	}
	checker := utils.HTTPBinUtil{}
	for {
		select {
		case pxy := <-ch.Recv():
			totalCount++
			go func() {
				if checker.ProxyHTTPSUsable(pxy.URL()) {
					fmt.Printf("%d\t%+v\n", totalCount, pxy)
				}
			}()
		}
	}
}
