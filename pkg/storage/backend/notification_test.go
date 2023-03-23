// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/leosocy/proksi/pkg/pubsub"
	"github.com/leosocy/proksi/pkg/storage"
)

func TestInsertionWatcher(t *testing.T) {
	backend := NewInMemoryBackend()
	nb := WithNotifier(backend, &pubsub.BaseNotifier{})
	recvCount := 0
	watcher := NewInsertionWatcher(func(pxy *proxy.Proxy) {
		recvCount++
	}, storage.FilterUptime(80))
	nb.Attach(watcher)

	pxy, _ := proxy.NewBuilder().AddrPort("1.2.3.4:80").Build()
	// insert a new proxy pass filters
	nb.Insert(pxy)
	// update a exists proxy
	nb.Update(pxy)
	pxy.Score = 80
	nb.InsertOrUpdate(pxy)
	// insert or update new proxy not pass filters
	anotherPxy, _ := proxy.NewBuilder().AddrPort("5.6.7.8:80").Build()
	anotherPxy.AddScore(-50)
	nb.InsertOrUpdate(anotherPxy)
	// delete a proxy
	nb.Delete(pxy)
	// assert only notify when insert pxy
	assert.Equal(t, 1, recvCount)
}
