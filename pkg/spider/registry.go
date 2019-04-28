// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

const (
	NameOfXici string = "xici"
)

func NewSpiderName(name string) *Spider {
	switch name {
	case NameOfXici:
		return NewSpider(
			name,
			Urls(buildXiciUrls()),
			XPathQuery(`//*[@id="ip_list"]/tbody/tr[@class='odd']/td[2]`),
		)
	}
	return nil
}
