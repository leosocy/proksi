// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package backend

import (
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
	"github.com/Leosocy/IntelliProxy/pkg/pubsub"
	"github.com/Leosocy/IntelliProxy/pkg/storage"
)

// NotifyBackend is an interface that notify when data changes in backend.
type NotifyBackend interface {
	pubsub.Notifier
	Backend
}

// Event represents a single backend notification.
type Event struct {
	Op  Op           // Backend operation that triggered the event.
	Pxy *proxy.Proxy // Related proxy.
}

// Op describes a set of backend operations.
type Op uint32

// These are the generalized backend operations that can trigger a notification.
const (
	Insert Op = iota
	Update
	Delete
)

// notifyBackendWrapper implements NotifyBackend interface.
// It wraps the Backend's `Insert/InsertOrUpdate` method to send event
// when the new proxy inserted, updated or deleted.
type notifyBackendWrapper struct {
	pubsub.Notifier
	Backend
}

func (nb *notifyBackendWrapper) Insert(p *proxy.Proxy) (err error) {
	if err = nb.Backend.Insert(p); err == nil {
		nb.Notify(&Event{Insert, p})
	}
	return
}

func (nb *notifyBackendWrapper) Update(newP *proxy.Proxy) (err error) {
	if err = nb.Backend.Update(newP); err == nil {
		nb.Notify(&Event{Update, newP})
	}
	return
}

func (nb *notifyBackendWrapper) InsertOrUpdate(p *proxy.Proxy) (inserted bool, err error) {
	if inserted, err = nb.Backend.InsertOrUpdate(p); err == nil {
		var op Op
		if inserted {
			op = Insert
		} else {
			op = Update
		}
		nb.Notify(&Event{op, p})
	}
	return
}

func (nb *notifyBackendWrapper) Delete(p *proxy.Proxy) (err error) {
	if err = nb.Backend.Delete(p); err == nil {
		nb.Notify(&Event{Delete, p})
	}
	return
}

// WithNotifier returns a notifiable backend with notifier
func WithNotifier(backend Backend, notifier pubsub.Notifier) NotifyBackend {
	return &notifyBackendWrapper{Notifier: notifier, Backend: backend}
}

// InsertionWatcher only interested in the new proxy inserted event,
// and will notify when the proxy passed filtered if filters set.
type InsertionWatcher struct {
	recvCh  chan<- *proxy.Proxy
	filters []storage.Filter
}

func NewInsertionWatcher(recvCh chan<- *proxy.Proxy, fn ...storage.Filter) *InsertionWatcher {
	return &InsertionWatcher{
		recvCh:  recvCh,
		filters: fn,
	}
}

// Receipt implements pubsub.Watcher interface with filters.
func (w *InsertionWatcher) Receipt(obj interface{}) {
	if e, ok := obj.(*Event); ok && e.Op == Insert {
		w.receipt(e)
	}
}

func (w *InsertionWatcher) receipt(e *Event) {
	proxies := []*proxy.Proxy{e.Pxy}
	for _, filter := range w.filters {
		proxies = filter(proxies)
	}
	if len(proxies) == 0 {
		return
	}
	w.recvCh <- proxies[0]
}
