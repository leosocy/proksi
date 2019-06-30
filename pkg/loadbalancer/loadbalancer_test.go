// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package loadbalancer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type simpleEndpoint struct {
	w int
}

func (e simpleEndpoint) Weight() int {
	return e.w
}

func (e simpleEndpoint) String() string {
	return fmt.Sprintf("%d", e.w)
}

var endpointsData = []Endpoint{
	simpleEndpoint{w: 1},
	simpleEndpoint{w: 2},
	simpleEndpoint{w: 4},
	simpleEndpoint{w: 8},
	simpleEndpoint{w: 16},
}

func TestEndpointsManipulate(t *testing.T) {
	assert := assert.New(t)
	eps := newEndpoints(endpointsData[:2]...)
	assert.Len(eps.store, 2)
	for _, e := range endpointsData {
		eps.add(e)
	}
	assert.Len(eps.store, len(endpointsData))
	eps.del(simpleEndpoint{8})
	assert.Len(eps.store, len(endpointsData)-1)
}

func BenchmarkEndpointsManipulate(b *testing.B) {
	eps := newEndpoints()
	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			eps.add(simpleEndpoint{w: i})
		}
	})
	b.Run("del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			eps.del(simpleEndpoint{w: i})
		}
	})
}

func TestLoadBalancerSelect(t *testing.T) {
	lbs := map[string]LoadBalancer{
		"Random":     NewLoadBalancer(Random, endpointsData...),
		"RoundRobin": NewLoadBalancer(RoundRobin, endpointsData...),
		"WeightedRoundRobin": NewLoadBalancer(WeightedRoundRobin, endpointsData...),
	}
	for name, lb := range lbs {
		stats := make(map[string]int)
		for i := 0; i < 5*len(endpointsData); i++ {
			ep := lb.Select()
			stats[ep.String()]++
		}
		fmt.Printf("%s %+v\n", name, stats)
	}
}
