// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package spider

import (
	"bytes"
	"os"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/Masterminds/sprig/v3"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/leosocy/proksi/pkg/proxy"
)

// ConfigurationError is the type of error returned from a constructor (e.g. NewSpider)
// when the specified configuration is invalid.
type ConfigurationError string

func (err ConfigurationError) Error() string {
	return "spider: invalid configuration (" + string(err) + ")"
}

type Config struct {
	// ID is the unique identifier of a Spider
	ID string
	// Name is the description of the Spider
	Name string
	// Enabled represents whether to enable Spider
	Enabled bool
	Urls    []string
	Parser  *ParserConfig
	Rule    struct {
		// Interval is the duration until next round full requests of urls
		Interval time.Duration
		Limit    struct {
			// Parallelism is the number of the maximum allowed concurrent requests of the matching domains
			Parallelism int
			// Delay is the duration to wait before creating a new request to the matching domains
			Delay time.Duration
			// Jitter is the extra randomized duration to wait added to Delay before creating a new request
			Jitter time.Duration
		}
	}
}

// Configure set some configurations with sane defaults.
func (c *Config) Configure() *Config {
	if c.Rule.Interval == 0 {
		c.Rule.Interval = 2 * time.Hour
	}
	if c.Rule.Limit.Parallelism == 0 {
		c.Rule.Limit.Parallelism = 2
	}
	if c.Rule.Limit.Delay == 0 {
		c.Rule.Limit.Delay = 10 * time.Second
	}
	if c.Rule.Limit.Jitter == 0 {
		c.Rule.Limit.Jitter = c.Rule.Limit.Delay / 10
	}
	return c
}

// Validate checks a Config instance. It will return a
// ConfigurationError if the specified values don't make sense.
func (c *Config) Validate() error {
	switch {
	case c.ID == "":
		return ConfigurationError("ID must be non-empty string")
	case c.Name == "":
		return ConfigurationError("Name must be non-empty string")
	}

	switch {
	case c.Rule.Interval < 10*time.Minute:
		return ConfigurationError("Rule.Interval must be >= 10min")
	case c.Rule.Limit.Parallelism < 1:
		return ConfigurationError("Rule.Limit.Parallelism must be >= 1")
	case c.Rule.Limit.Delay < 5*time.Second:
		return ConfigurationError("Rule.Limit.Delay must be >= 5s")
	}
	log.Debug().Str("spider", c.ID).Msgf("valid config %+v", c)

	return nil
}

type Spider struct {
	config *Config

	c      *colly.Collector
	parser ProxyParser
	pc     proxy.Collector
	logger zerolog.Logger

	round     uint32
	startOnce sync.Once
	roundChan chan struct{}
	stopChan  chan struct{}
}

func (s *Spider) crawlOnce() {
	atomic.AddUint32(&s.round, 1)
	s.logger.Debug().Uint32("round", s.round).Msg("start crawling")

	for _, url := range s.config.Urls {
		if err := s.c.Visit(url); err != nil {
			s.logger.Warn().Uint32("round", s.round).Str("url", url).Err(err).Msg("Failed to crawl")
		}
	}
}

func (s *Spider) nextRound(d time.Duration) {
	time.AfterFunc(d, func() {
		s.roundChan <- struct{}{}
	})
}

func (s *Spider) Start() {
	if !s.config.Enabled {
		s.logger.Warn().Msg("The spider is disabled, nothing happened.")
		return
	}
	s.startOnce.Do(func() {
		go func() {
			s.nextRound(0)
			for {
				select {
				case <-s.roundChan:
					s.crawlOnce()
					s.nextRound(s.config.Rule.Interval)
				case <-s.stopChan:
					s.stopCrawl()
				}
			}
		}()
	})
}

func (s *Spider) stopCrawl() {
	close(s.roundChan)
	close(s.stopChan)
}

func (s *Spider) Stop() {
	s.stopChan <- struct{}{}
}

func NewSpider(config *Config) (*Spider, error) {
	config.Configure()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	parser, err := NewProxyParser(config.ID, config.Parser)
	if err != nil {
		return nil, err
	}

	c := colly.NewCollector(
		colly.Async(false),
		colly.UserAgent(browser.Random()),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
	)
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       config.Rule.Limit.Delay,
		RandomDelay: config.Rule.Limit.Jitter,
		Parallelism: config.Rule.Limit.Parallelism,
	}); err != nil {
		return nil, err
	}
	pc := proxy.LogCollector{}

	s := &Spider{
		config:    config,
		c:         c,
		parser:    parser,
		pc:        pc,
		logger:    zerolog.New(os.Stderr).With().Str("module", "spider").Str("spider", config.ID).Logger(),
		round:     0,
		startOnce: sync.Once{},
		roundChan: make(chan struct{}, 1),
		stopChan:  make(chan struct{}, 1),
	}

	s.c.OnResponse(func(response *colly.Response) {
		s.logger.Debug().Uint32("round", s.round).Str("url", response.Request.URL.String()).Msg("Crawl done")
		parser.HandleResponse(response, pc)
	})
	s.c.OnError(func(response *colly.Response, err error) {
		s.logger.Warn().Uint32("round", s.round).Str("url", response.Request.URL.String()).Err(err).Msg("")
	})

	return s, nil
}

type Values map[string]interface{}

func RenderSpiders(tplPath string, values Values) ([]*Spider, error) {
	spiders := make([]*Spider, 0, 8)
	b, err := os.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("spiders").Funcs(sprig.FuncMap()).Parse(string(b))
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, values)
	if err != nil {
		return nil, err
	}
	var spiderConfigs map[string]*Config
	err = yaml.Unmarshal(buf.Bytes(), &spiderConfigs)
	if err != nil {
		return nil, err
	}
	for id, config := range spiderConfigs {
		config.ID = id
		spd, err := NewSpider(config)
		if err != nil {
			continue
		}
		spiders = append(spiders, spd)
	}
	return spiders, nil
}
