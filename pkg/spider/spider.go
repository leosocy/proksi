// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"sync/atomic"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Sirupsen/logrus"
	"github.com/gocolly/colly"

	"github.com/leosocy/proksi/pkg/proxy"
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

const (
	Idle = iota
	Crawling
	CoolDown
)

// Spider provides the instance for crawling jobs.
type Spider struct {
	name      string
	parser    spiderCoreParser
	c         *colly.Collector
	ch        proxy.CachedChan
	logger    *logrus.Entry
	period    time.Duration
	coolDown  time.Duration
	state     uint32
	needCrawl chan bool
}

func newSpider(name string, parser spiderCoreParser, options ...func(*Spider)) *Spider {
	s := &Spider{name: name, parser: parser, state: Idle, needCrawl: make(chan bool)}
	s.init()

	for _, opt := range options {
		opt(s)
	}

	s.registerCallbacks()

	return s
}

// Limit sets the rule used by the Collector.
func Limit(rule *colly.LimitRule) func(*Spider) {
	return func(s *Spider) {
		if err := s.c.Limit(rule); err != nil {
			panic(err)
		}
	}
}

// Period sets the interval duration, which is used
// to set the sleep time after each url is crawled.
func Period(d time.Duration) func(*Spider) {
	return func(s *Spider) {
		s.period = d
	}
}

// CoolDownTime sets the sleep time after crawlOnce,
// the purpose is to reduce the risk being banned of ip by the website.
func CoolDownTime(d time.Duration) func(*Spider) {
	return func(s *Spider) {
		s.coolDown = d
	}
}

// Init initializes the Spider's private variables
// and sets default configuration for the Spider
func (s *Spider) init() {
	s.c = colly.NewCollector(
		colly.Async(false),
		colly.UserAgent(browser.Random()),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
	)
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	s.logger = logger.WithFields(logrus.Fields{
		"spider-name": s.name,
	})
}

// registerCallbacks registers some callbacks after option spider.
func (s *Spider) registerCallbacks() {
	s.c.OnRequest(func(r *colly.Request) {
		s.logger.Infof("Crawling %s", r.URL)
	})

	s.c.OnResponse(func(r *colly.Response) {
		s.logger.Infof("Crawl %s done", r.Request.URL)
	})

	s.c.OnXML(s.parser.Query(), func(e *colly.XMLElement) {
		ip, port := s.parser.Parse(e)
		if s.ch != nil {
			s.ch.Send(ip, port)
		} else {
			s.logger.Infof("%s:%s\n", ip, port)
		}
	})
}

// crawlOnce traverses urls and visit for each url and send it to cached channel.
func (s *Spider) crawlOnce() {
	if !atomic.CompareAndSwapUint32(&s.state, Idle, Crawling) {
		return
	}
	s.logger.Info("Start crawling once")
	for _, url := range s.parser.Urls() {
		if err := s.c.Visit(url); err != nil {
			s.logger.Warnf("Failed to crawl %s, %v", url, err)
		}
	}
	s.logger.Info("Finish crawling once")
	atomic.CompareAndSwapUint32(&s.state, Crawling, CoolDown)
	s.logger.Infof("Enter %f s cool down time", s.coolDown.Seconds())
	time.Sleep(s.coolDown)
	atomic.CompareAndSwapUint32(&s.state, CoolDown, Idle)
}

// TryCrawl sends object to needCrawl chan when this spider id IDLE.
func (s *Spider) TryCrawl() {
	if atomic.LoadUint32(&s.state) == Idle {
		s.needCrawl <- true
	}
}

// Start calls crawlOnce after sleeping the period duration or when receiving a re-crawl chan.
func (s *Spider) Start(ch proxy.CachedChan) {
	s.ch = ch
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()
	for {
		select {
		case <-s.needCrawl:
			s.crawlOnce()
		case <-ticker.C:
			s.crawlOnce()
		}
	}
}
