// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package pubsub

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/Leosocy/IntelliProxy/mocks"
	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

func TestBaseNotifier(t *testing.T) {
	pxy, _ := proxy.NewProxy("1.2.3.4", "80")
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
