// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"net/http"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/PuerkitoBio/goquery"
)

const xiciBaseURL = "https://www.xicidaili.com"

type xiciParser struct{}

func (p xiciParser) Parse(response *http.Response, proxyCh chan<- *proxy.Proxy) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return
	}
	doc.Find("table").Find("tr").Each(func(_ int, tr *goquery.Selection) {
		if tr.HasClass("odd") || tr.HasClass("") {
			ip := tr.Children().Get(1).FirstChild.Data
			port := tr.Children().Get(2).FirstChild.Data
			if proxy, err := proxy.NewProxy(ip, port); err == nil {
				proxyCh <- proxy
			}
		}
	})
}

func buildXiciUrls() (urls []string) {
	for _, domain := range []string{"nn", "nt", "wn", "wt"} {
		for page := 1; page <= 10; page++ {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", xiciBaseURL, domain, page))
		}
	}
	return
}
