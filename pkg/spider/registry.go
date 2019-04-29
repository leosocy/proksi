// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"

	"github.com/gocolly/colly"
)

const (
	// NameOfXici 西刺代理 `https://www.xicidaili.com/nn/`
	NameOfXici = "xici"
	// NameOfKuai 快代理 `https://www.kuaidaili.com/ops/`, `https://www.kuaidaili.com/free/`
	NameOfKuai = "kuai"
	// NameOfYun 云代理，质量较高. `http://www.ip3366.net/free/`
	NameOfYun = "yun"
	// NameOfIphai ip海代理，`http://www.iphai.com/free/ng`
	NameOfIphai = "iphai"
)

// NewSpider creates a new Spider with name and default configurations.
func NewSpider(name string) *Spider {
	switch name {
	case NameOfXici:
		return newSpider(name, xiciSpider{})
	case NameOfKuai:
		return newSpider(name, kuaiSpider{})
	case NameOfYun:
		return newSpider(name, yunSpider{})
	case NameOfIphai:
		return newSpider(name, iphaiSpider{})
	default:
		return nil
	}
}

type xiciSpider struct{}

func (s xiciSpider) Urls() (urls []string) {
	baseURL := "https://www.xicidaili.com"
	for page := 1; page <= 10; page++ {
		for _, domain := range []string{"nn", "nt", "wn", "wt"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", baseURL, domain, page))
		}
	}
	return
}
func (s xiciSpider) Query() string {
	return `//*[@id="ip_list"]/tbody/tr[@class="odd" or ""]`
}
func (s xiciSpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[2]")
	port = e.ChildText("td[3]")
	return
}

type kuaiSpider struct{}

func (s kuaiSpider) Urls() (urls []string) {
	baseURL := "https://www.kuaidaili.com/free"
	freeBaseURL := "https://www.kuaidaili.com/ops/proxylist"
	for page := 1; page <= 10; page++ {
		urls = append(urls, fmt.Sprintf("%s/%d", freeBaseURL, page))
		for _, domain := range []string{"intr", "inha"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", baseURL, domain, page))
		}
	}
	return
}
func (s kuaiSpider) Query() string {
	return `//*[@id="list" or "freelist"]/table/tbody/tr`
}
func (s kuaiSpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[@data-title='IP']")
	port = e.ChildText("td[@data-title='PORT']")
	return
}

type yunSpider struct{}

func (s yunSpider) Urls() (urls []string) {
	yunBaseURL := "http://www.ip3366.net/free"
	for page := 1; page <= 7; page++ {
		for stype := 1; stype <= 2; stype++ {
			urls = append(urls, fmt.Sprintf("%s/?stype=%d&page=%d", yunBaseURL, stype, page))
		}
	}
	return
}
func (s yunSpider) Query() string {
	return `//*[@id="list"]/table/tbody/tr`
}
func (s yunSpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[1]")
	port = e.ChildText("td[2]")
	return
}

type iphaiSpider struct{}

func (s iphaiSpider) Urls() (urls []string) {
	baseURL := "http://www.iphai.com/free"
	for _, domain := range []string{"ng", "np", "wg", "wp"} {
		urls = append(urls, fmt.Sprintf("%s/%s", baseURL, domain))
	}
	return urls
}
func (s iphaiSpider) Query() string {
	return `/html/body/div[2]/div[2]/table/tbody/tr[position()>1]`
}
func (s iphaiSpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[1]")
	port = e.ChildText("td[2]")
	return
}
