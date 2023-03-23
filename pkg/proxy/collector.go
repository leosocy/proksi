// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Collector defines the interface for collecting proxies.
type Collector interface {
	// Collect method is used to collect proxies.
	Collect(ps ...*Proxy)
	// Close method is used to close the collector.
	Close() error
}

type logCollector struct{}

// NewLogCollector creates a simple collector that logs collected proxies.
func NewLogCollector() Collector {
	return logCollector{}
}

func (c logCollector) Collect(ps ...*Proxy) {
	for _, p := range ps {
		log.Info().Msgf("collect proxy %s", p.AddrPort.String())
	}
}

func (c logCollector) Close() error {
	return nil
}

type batchedCollector struct {
	child     Collector
	proxies   []*Proxy
	proxyChan chan *Proxy

	waitTime  time.Duration
	batchSize int

	closeOnce sync.Once
	closed    chan struct{}
	logger    zerolog.Logger
}

// NewBatchedCollector creates a new batch collector that batches collected proxies and sends to a child Collector.
// It accumulates proxies for a certain amount of time or until a certain number of proxies are collected,
// then flushes them to the child Collector. with the given child Collector, wait time, and batch size.
func NewBatchedCollector(child Collector, waitTime time.Duration, batchSize int) Collector {
	c := &batchedCollector{
		child:     child,
		proxies:   make([]*Proxy, 0, batchSize),
		proxyChan: make(chan *Proxy),
		waitTime:  waitTime,
		batchSize: batchSize,
		closeOnce: sync.Once{},
		closed:    make(chan struct{}),
		logger:    zerolog.New(os.Stderr).With().Str("module", "proxy").Str("collector", "batched").Logger(),
	}
	go c.accumulate()
	return c
}

func (c *batchedCollector) flush() {
	if len(c.proxies) > 0 {
		c.child.Collect(c.proxies...)
		c.proxies = c.proxies[:1]
		c.logger.Debug().Msgf("flushed %d proxies", len(c.proxies))
	}
}

func (c *batchedCollector) accumulate() {
	defer close(c.closed)

	ticker := time.NewTicker(c.waitTime)
	defer ticker.Stop()

	for {
		select {
		case pxy, ok := <-c.proxyChan:
			if ok {
				c.proxies = append(c.proxies, pxy)
				if len(c.proxies) >= c.batchSize {
					c.flush()
				}
			} else {
				c.flush()
				c.logger.Info().Msg("exit accumulation")
				return
			}
		case <-ticker.C:
			c.flush()
		}
	}
}

func (c *batchedCollector) Collect(ps ...*Proxy) {
	for _, p := range ps {
		c.proxyChan <- p
	}
}

func (c *batchedCollector) Close() error {
	var err error
	c.closeOnce.Do(func() {
		// close proxyChan and wait until the background accumulate goroutine recv all data in chan and exit.
		close(c.proxyChan)
		<-c.closed

		if e := c.child.Close(); e != nil {
			err = e
		}
	})
	return err
}
