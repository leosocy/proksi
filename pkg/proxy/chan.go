// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"hash/fnv"

	"github.com/steakknife/bloomfilter"
)

// CachedChan provides a channel to transport proxies from spiders.
type CachedChan interface {
	Send(ip, port string)
	Recv() <-chan *Proxy
}

// NewBloomCachedChan returns a default bloom cached chan.
func NewBloomCachedChan() CachedChan {
	bf, err := bloomfilter.NewOptimal(1024*1024, 0.000000001)
	if err != nil {
		panic(err)
	}
	return &BloomCachedChan{
		entryBf: bf,
		ch:      make(chan *Proxy, 1024),
	}
}

// BloomCachedChan excludes proxy that are already sent to channel
// by placing a bloom filter in front of the channel.
type BloomCachedChan struct {
	// entryBf is a bloomfilter that determines
	// whether the proxy has been added to the channel.
	entryBf *bloomfilter.Filter
	// ch transports proxies that crawled by spiders.
	ch chan *Proxy
}

func (cc *BloomCachedChan) Send(ip, port string) {
	if pxy, err := NewBuilder().IP(ip).Port(port).Build(); err == nil {
		hasher := fnv.New64()
		if _, err := hasher.Write(pxy.IP); err == nil &&
			!cc.entryBf.Contains(hasher) {
			// first add it to filter, since send to
			// channel will block current goroutine.
			cc.entryBf.Add(hasher)
			cc.ch <- pxy
		}
	}
}

func (cc *BloomCachedChan) Recv() <-chan *Proxy {
	return cc.ch
}
