// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package loadbalancer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Endpoint interface {
	Weight() int
	fmt.Stringer
}

type endpoints struct {
	store   []Endpoint
	indices map[Endpoint]int
	mu      sync.RWMutex
}

func newEndpoints(data ...Endpoint) *endpoints {
	es := &endpoints{
		store:   make([]Endpoint, 0, len(data)),
		indices: make(map[Endpoint]int, len(data)),
	}
	for i, e := range data {
		es.store = append(es.store, e)
		es.indices[e] = i
	}
	return es
}

func (es *endpoints) add(e Endpoint) {
	es.mu.Lock()
	defer es.mu.Unlock()
	if _, ok := es.indices[e]; ok {
		return
	}
	es.store = append(es.store, e)
	es.indices[e] = len(es.store) - 1
}

func (es *endpoints) del(e Endpoint) {
	es.mu.Lock()
	defer es.mu.Unlock()
	if idx, ok := es.indices[e]; ok {
		copy(es.store[idx:], es.store[idx+1:])
		es.store = es.store[:len(es.store)-1]
		delete(es.indices, e)
	}
}

type Strategy uint8

const (
	Random Strategy = iota
	RoundRobin
	WeightedRoundRobin
)

type LoadBalancer interface {
	Select() Endpoint
}

type base struct {
	es *endpoints
}

func (lb *base) AddEndpoint(e Endpoint) {
	lb.es.add(e)
}

func (lb *base) DelEndpoint(e Endpoint) {
	lb.es.del(e)
}

// NewLoadBalancer returns the specified load balancer
// according to strategy, and initializes with endpoints
//
// Random: The random load balancer forwards a client request randomly.
// RoundRobin: The round-robin load balancer forwards a client request to each endpoint in turn.
// WeightedRoundRobin: In the weighted round-robin algorithm, each endpoints is assigned a value that signifies,
// relative to the other endpoints in the pool, how that endpoint performs.
func NewLoadBalancer(strategy Strategy, endpoints ...Endpoint) LoadBalancer {
	switch strategy {
	case Random:
		return &random{
			&base{es: newEndpoints(endpoints...)},
		}
	case RoundRobin:
		return &roundRobin{
			rand.Uint64(),
			&base{es: newEndpoints(endpoints...)},
		}
	case WeightedRoundRobin:
		return &weightedRoundRobin{
			make(map[Endpoint]int),
			&base{es: newEndpoints(endpoints...)},
		}
	default:
		panic("unknown load balance strategy")
	}
}

type random struct {
	*base
}

func (lb *random) Select() Endpoint {
	lb.es.mu.RLock()
	defer lb.es.mu.RUnlock()
	endpoints := lb.es.store
	if len(endpoints) == 0 {
		return nil
	}
	return endpoints[rand.Int()%len(endpoints)]
}

type roundRobin struct {
	cursor uint64
	*base
}

func (lb *roundRobin) Select() Endpoint {
	lb.es.mu.RLock()
	defer lb.es.mu.RUnlock()
	endpoints := lb.es.store
	if len(endpoints) == 0 {
		return nil
	}
	defer atomic.AddUint64(&lb.cursor, 1)
	return endpoints[lb.cursor%uint64(len(endpoints))]
}

type weightedRoundRobin struct {
	endpointsCurrentWeight map[Endpoint]int
	*base
}

func (lb *weightedRoundRobin) Select() Endpoint {
	lb.es.mu.RLock()
	defer lb.es.mu.RUnlock()
	endpoints := lb.es.store
	if len(endpoints) == 0 {
		return nil
	}
	totalWeight := 0
	var endpointOfMaxWeight Endpoint
	for _, e := range endpoints {
		totalWeight += e.Weight()
		lb.endpointsCurrentWeight[e] += e.Weight()
		if endpointOfMaxWeight == nil ||
			lb.endpointsCurrentWeight[e] > lb.endpointsCurrentWeight[endpointOfMaxWeight] {
			endpointOfMaxWeight = e
		}
	}
	lb.endpointsCurrentWeight[endpointOfMaxWeight] -= totalWeight
	return endpointOfMaxWeight
}
