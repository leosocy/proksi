// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"time"

	"github.com/Leosocy/gipp/pkg/proxy"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Sirupsen/logrus"
	"github.com/gocolly/colly"
)

type spiderCoreParser interface {
	// Urls 返回spider要爬取的所有url
	Urls() []string
	// Query 用于找到爬取的XML中一条代理记录tr(子节点有td存储ip和port)，是一个xpath表达式，
	// 会注册到OnXML回调
	Query() string
	// Parse 用于解析通过Query找到的那条记录中的ip和port，会注册到OnXML的回调
	Parse(e *colly.XMLElement) (ip, port string)
}

// Spider provides the instance for crawling jobs.
type Spider struct {
	name   string
	parser spiderCoreParser
	c      *colly.Collector
	ch     proxy.CachedChan
	logger *logrus.Logger
}

func newSpider(name string, parser spiderCoreParser, options ...func(*Spider)) *Spider {
	s := &Spider{name: name, parser: parser}
	s.init()

	for _, opt := range options {
		opt(s)
	}

	s.registerCallbacks()

	return s
}

// Init initializes the Spider's private variables
// and sets default configuration for the Spider
func (s *Spider) init() {
	s.c = colly.NewCollector(
		colly.Async(false),
		colly.UserAgent(browser.Random()),
		colly.MaxDepth(1),
	)
	s.c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1, // avoids detection that I'm a spider
		Delay:       10 * time.Second,
	})

	s.logger = logrus.New()
	s.logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
}

// registerCallbacks registers some callbacks after option spider.
func (s *Spider) registerCallbacks() {
	entry := s.logger.WithFields(logrus.Fields{
		"spider-name": s.name,
	})

	s.c.OnRequest(func(r *colly.Request) {
		entry.Infof("start crawling %s.", r.URL)
	})

	s.c.OnResponse(func(r *colly.Response) {
		entry.Infof("crawl %s done.", r.Request.URL)
	})

	s.c.OnXML(s.parser.Query(), func(e *colly.XMLElement) {
		ip, port := s.parser.Parse(e)
		if s.ch != nil {
			s.ch.Send(ip, port)
		} else {
			fmt.Printf("%s:%s\n", ip, port)
		}
	})

	s.c.OnError(func(r *colly.Response, err error) {
		entry.Errorf("crawl %s failed. %v", r.Request.URL, err)
	})
}

// CrawlTo traverses urls and visit for
// each url and send it to cached channel.
func (s *Spider) CrawlTo(ch proxy.CachedChan) {
	s.ch = ch
	for _, url := range s.parser.Urls() {
		s.c.Visit(url)
	}
}
