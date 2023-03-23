// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"net"
	"testing"

	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/leosocy/proksi/pkg/quality"
	"github.com/stretchr/testify/assert"
)

func TestFilterScore(t *testing.T) {
	assert := assert.New(t)
	testData := []struct {
		proxies   []*proxy.Proxy
		threshold int8
		count     int
	}{
		{
			proxies: []*proxy.Proxy{
				{IP: net.ParseIP("1.1.1.1"), Port: 8000, Score: 30},
				{IP: net.ParseIP("2.2.2.2"), Port: 8000, Score: 50},
			},
			threshold: 40,
			count:     1,
		},
		{
			proxies: []*proxy.Proxy{
				{IP: net.ParseIP("1.1.1.1"), Port: 8000, Score: 30},
				{IP: net.ParseIP("2.2.2.2"), Port: 8000, Score: 50},
			},
			threshold: 60,
			count:     0,
		},
	}
	for _, data := range testData {
		filter := FilterUptime(quality.Uptime(data.threshold))
		proxies := filter(data.proxies)
		assert.Equal(data.count, len(proxies))
	}
}
