// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"

	"github.com/gocolly/colly"
)

const (
	NameOfXici = "xici"
	NameOfKuai = "kuai"
)

func NewSpiderName(name string) *Spider {
	switch name {
	case NameOfXici:
		return NewSpider(
			name,
			buildXiciUrls(),
			WrappedXMLCallback(
				`//*[@id="ip_list"]/tbody/tr[@class='odd' or '']`,
				func(e *colly.XMLElement) (ip, port string) {
					ip = e.ChildText("td[2]")
					port = e.ChildText("td[3]")
					return
				},
			),
		)
	case NameOfKuai:
		return NewSpider(
			name,
			buildKuaiUrls(),
			WrappedXMLCallback(
				`//*[@id="list"]/table/tbody/tr`,
				func(e *colly.XMLElement) (ip, port string) {
					ip = e.ChildText("td[@data-title='IP']")
					port = e.ChildText("td[@data-title='PORT']")
					return
				},
			),
		)
	}
	return nil
}

func buildXiciUrls() (urls []string) {
	xiciBaseURL := "https://www.xicidaili.com"
	for page := 1; page <= 10; page++ {
		for _, domain := range []string{"nn", "nt", "wn", "wt"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", xiciBaseURL, domain, page))
		}
	}
	return
}

func buildKuaiUrls() (urls []string) {
	kuaiBaseURL := "https://www.kuaidaili.com/free"
	for page := 1; page <= 10; page++ {
		for _, domain := range []string{"intr", "inha"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", kuaiBaseURL, domain, page))
		}
	}
	return
}
