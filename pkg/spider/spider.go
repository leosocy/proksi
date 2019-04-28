// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Sirupsen/logrus"
	"github.com/gocolly/colly"
)

type xmlCallbackContainer struct {
	Query    string
	Function func(*colly.XMLElement)
}

// Spider provides the instance for crawling jobs.
type Spider struct {
	name        string
	urls        []string
	xmlCallback xmlCallbackContainer
	c           *colly.Collector
	logger      *logrus.Entry
}

// NewSpider creates a new Spider instance with default configuration.
func NewSpider(name string, urls []string, options ...func(*Spider)) *Spider {
	s := &Spider{name: name, urls: urls}
	s.Init()

	for _, opt := range options {
		opt(s)
	}

	s.RegisterCallbacks()

	return s
}

// WrappedXMLCallback wraps the spider's xmlCallbackContainer
// with a callback which returns the parsed ip and port.
//
// The wrapped callback function adds the proxy to the ProxyPool.
func WrappedXMLCallback(query string,
	callback func(*colly.XMLElement) (ip, port string)) func(*Spider) {
	return func(s *Spider) {
		s.xmlCallback.Query = query
		s.xmlCallback.Function = func(e *colly.XMLElement) {
			ip, port := callback(e)
			fmt.Printf("%s:%s\n", ip, port)
		}
	}
}

// Init initializes the Spider's private variables
// and sets default configuration for the Spider
func (s *Spider) Init() {
	s.c = colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(browser.Random()),
	)
	s.c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       5 * time.Second,
	})

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	s.logger = logrus.WithFields(logrus.Fields{
		"package": "spider",
		"name":    s.name,
	})
}

// RegisterCallbacks registers some callbacks after option spider.
func (s *Spider) RegisterCallbacks() {
	s.c.OnRequest(func(r *colly.Request) {
		s.logger.Infof("start crawling %s", r.URL)
	})

	s.c.OnResponse(func(r *colly.Response) {
		s.logger.Infof("crawl %s done", r.Request.URL)
	})

	if s.xmlCallback.Function != nil {
		s.c.OnXML(s.xmlCallback.Query, s.xmlCallback.Function)
	}

	s.c.OnError(func(r *colly.Response, err error) {
		s.logger.Errorf("crawl %s failed. %v", r.Request.URL, err)
	})
}

// Crawl traverses urls and visit for each url.
func (s *Spider) Crawl() {
	for _, url := range s.urls {
		s.c.Visit(url)
	}
}
