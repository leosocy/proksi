// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import "github.com/Leosocy/gipp/pkg/proxy"

// Spider interface of all spiders.
type Spider interface {
	Crawl(chan<- *proxy.Proxy)
}
