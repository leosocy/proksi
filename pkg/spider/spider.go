// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"log"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/gocolly/colly"
)

// Spider provides the instance for crawling jobs.
type Spider struct {
	// Name is the name of spider defined in `spider/registry.go`
	Name          string
	urls          []string
	xpathQuery    string
	selectorQuery string
	proxyPool     chan<- *proxy.Proxy
	c             *colly.Collector
}

// NewSpider creates a new Spider instance with default configuration.
func NewSpider(name string, options ...func(*Spider)) *Spider {
	s := &Spider{Name: name}
	s.Init()

	for _, opt := range options {
		opt(s)
	}

	s.RegisterCallbacks()

	return s
}

// Urls sets the urls of the spider.
func Urls(urls []string) func(*Spider) {
	return func(s *Spider) {
		s.urls = urls
	}
}

// XPathQuery sets the query that used to locate the ip and port in XML.
func XPathQuery(xpath string) func(*Spider) {
	return func(s *Spider) {
		s.xpathQuery = xpath
	}
}

// SelectorQuery sets the query that used to locate the ip and port in HTML.
// If not set, spider won't handle OnHTML.
func SelectorQuery(selector string) func(*Spider) {
	return func(s *Spider) {
		s.selectorQuery = selector
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
		Parallelism: 4,
		Delay:       5 * time.Second,
	})
}

// RegisterCallbacks registers some callbacks after option spider.
func (s *Spider) RegisterCallbacks() {
	if s.xpathQuery != "" {
		s.c.OnXML(s.xpathQuery, func(e *colly.XMLElement) {
			fmt.Printf("%+v", e)
		})
	}
	if s.selectorQuery != "" {
		s.c.OnHTML(s.selectorQuery, func(e *colly.HTMLElement) {
			fmt.Printf("%+v", e)
		})
	}
	s.c.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})
}

func (s *Spider) Start() {
	for _, url := range s.urls {
		s.c.Visit(url)
	}
}
