// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"errors"
	"github.com/Leosocy/IntelliProxy/pkg/pubsub"
	"net"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

// Errors occur when using backend.
var (
	ErrProxyInvalid       = errors.New("proxy is nil or score <= 0")
	ErrProxyDuplicated    = errors.New("proxy is already in backend")
	ErrProxyDoesNotExists = errors.New("proxy doesn't exists")
	ErrProxyNoneAvailable = errors.New("proxy none available")
)

// Iterator is the function which will be call for each proxy in backend.
// It will stop when the iterator returns false.
type Iterator func(pxy *proxy.Proxy) bool

// Backend is an interface that store and manipulate proxies
type Backend interface {
	Insert(p *proxy.Proxy) error
	Update(newP *proxy.Proxy) error
	InsertOrUpdate(p *proxy.Proxy) (inserted bool, err error)
	Delete(ip net.IP) error
	Search(ip net.IP) *proxy.Proxy
	// Select returns proxies after filter with options
	Select(opts ...SelectOption) ([]*proxy.Proxy, error)
	Len() uint
	// TopK returns the first K proxies order by score descend.
	// If k is equal to 0, return all proxies in the backend
	TopK(k int) []*proxy.Proxy
	Iter(iter Iterator)
}

// BackendNotifier is an interface that notify watchers when data changes in backend.
type BackendNotifier interface {
	Backend
	pubsub.Notifier
}

// notifiableBackend implements BackendNotifier interface.
// It wraps the Backend's `Insert/InsertOrUpdate` method to send Notify when the new proxy inserted.
type notifiableBackend struct {
	Backend
	pubsub.Notifier
}

func (nb *notifiableBackend) Insert(p *proxy.Proxy) (err error) {
	if err = nb.Backend.Insert(p); err == nil {
		nb.Notify(p)
	}
	return
}

func (nb *notifiableBackend) InsertOrUpdate(p *proxy.Proxy) (inserted bool, err error) {
	if inserted, err = nb.Backend.InsertOrUpdate(p); inserted && err == nil {
		nb.Notify(p)
	}
	return
}

// WithNotifier returns a notifiable backend with notifier
func WithNotifier(backend Backend, notifier pubsub.Notifier) BackendNotifier {
	return &notifiableBackend{backend, notifier}
}

type BaseWatcher struct {
	recvCh  chan<- *proxy.Proxy
	filters []Filter
}

func NewBaseWatcher(recvCh chan<- *proxy.Proxy, fn ...Filter) *BaseWatcher {
	return &BaseWatcher{
		recvCh:  recvCh,
		filters: fn,
	}
}

// Update implements pubsub.Watcher interface with filters.
func (w *BaseWatcher) Update(obj interface{}) {
	pxy := obj.(*proxy.Proxy)
	proxies := []*proxy.Proxy{pxy}
	for _, filter := range w.filters {
		proxies = filter(proxies)
	}
	if len(proxies) == 0 {
		return
	}
	w.recvCh <- proxies[0]
}
