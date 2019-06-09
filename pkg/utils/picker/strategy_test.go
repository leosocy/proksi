// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package picker

import (
	"net"
	"testing"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

func TestStrategies(t *testing.T) {
	testData := []interface{}{
		&proxy.Proxy{IP: net.ParseIP("1.1.1.1"), Port: 8000, Score: 30},
		&proxy.Proxy{IP: net.ParseIP("2.2.2.2"), Port: 8000, Score: 50},
	}
	for name, strategy := range map[string]Strategy{"random": Random, "roundrobin": RoundRobin} {
		next := strategy(testData)
		counts := make(map[string]int)

		for i := 0; i < 100; i++ {
			elem, err := next()
			if err != nil {
				t.Fatal(err)
			}
			counts[elem.(*proxy.Proxy).URL()]++
		}

		t.Logf("%s: %+v\n", name, counts)
	}
}
