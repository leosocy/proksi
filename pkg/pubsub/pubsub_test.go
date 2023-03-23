// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package pubsub

import (
	"testing"

	"github.com/leosocy/proksi/mocks"
	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/stretchr/testify/mock"
)

func TestBaseNotifier(t *testing.T) {
	pxy, _ := proxy.NewBuilder().AddrPort("1.2.3.4:80").Build()
	notifier := BaseNotifier{}
	watcher := new(mocks.Watcher)
	watcher.On("Receipt", mock.Anything).Return()
	// notify with empty watchers
	notifier.Notify(pxy)
	watcher.AssertNotCalled(t, "Receipt", pxy)
	// attach
	notifier.Attach(watcher)
	notifier.Notify(pxy)
	watcher.AssertNumberOfCalls(t, "Receipt", 1)
	// attach another
	anotherWatcher := new(mocks.Watcher)
	anotherWatcher.On("Receipt", mock.Anything).Return()
	notifier.Attach(anotherWatcher)
	notifier.Notify(pxy)
	watcher.AssertNumberOfCalls(t, "Receipt", 2)
	anotherWatcher.AssertNumberOfCalls(t, "Receipt", 1)
	// detach all
	notifier.Detach(watcher)
	notifier.Detach(anotherWatcher)
	notifier.Notify(pxy)
	watcher.AssertNumberOfCalls(t, "Receipt", 2)
	anotherWatcher.AssertNumberOfCalls(t, "Receipt", 1)
}
