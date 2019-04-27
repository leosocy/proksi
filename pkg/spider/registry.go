// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

const (
	NameOfXici string = "xici"
)

// NewSpider returns a new spider accord to the name.
func NewSpider(name string) *Spider {
	switch name {
	case NameOfXici:
		return &Spider{
			name: name,
			urls: buildXiciUrls(),
			c:    DefaultCrawler{},
			p:    xiciParser{},
		}
	default:
		return nil
	}
}
