// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"errors"
	"net/http"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
)

// Crawler is the interface that wraps the basic Crawl method.
//
// It crawls a url of proxy website, and returns the response and error.
type Crawler interface {
	Crawl(url string) (response *http.Response, err error)
}

// DefaultCrawler implements Crawl.
type DefaultCrawler struct{}

func (c DefaultCrawler) Crawl(url string) (response *http.Response, err error) {
	logrus.Infof("[spider][default crawler] start crawling %s", url)
	resp, _, errs := gorequest.New().Proxy("http://111.177.171.189:9999").
		Get(url).Set("User-Agent", browser.Random()).EndBytes()
	if errs != nil {
		err = errs[0]
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		err = errors.New("crawl failed, response is nil or status not ok")
	}
	if err != nil {
		logrus.Warnf("[spider][default crawler] crawl %s failed. %v", url, err)
	}
	return resp, err
}

// Parser is the interface that wraps the basic Parse method.
//
// It parse the http response from a proxy website
// to a html document or something others,
// and add the proxy records to a channel.
type Parser interface {
	Parse(response *http.Response, proxyCh chan<- *proxy.Proxy)
}

// Spider crawl using C and parse response using R
type Spider struct {
	name string
	urls []string
	c    Crawler
	p    Parser
}

// Do iterates over the urls and then calls c.rawl and p.parse,
// respectively, to add the proxy to the channel
func (s *Spider) Do(proxyCh chan<- *proxy.Proxy) {
	for _, url := range s.urls {
		if resp, err := s.c.Crawl(url); err == nil {
			s.p.Parse(resp, proxyCh)
		}
	}
}
