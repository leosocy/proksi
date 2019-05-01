// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"strings"

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
	// NameOfXila 西拉免费代理，`http://www.xiladaili.com/`
	NameOfXila = "xila"
	// NameOfNima 泥马代理，量较大，`http://www.nimadaili.com/`
	NameOfNima = "nima"
	// NameOfEightnine 89免费代理，`http://www.89ip.cn/`
	NameOfEightnine = "eightnine"
	// NameOfHappy 开心代理，`http://ip.kxdaili.com/`
	NameOfHappy = "kaixin"
)

// BuildAndInitAll returns all of the enable spider.
func BuildAndInitAll() (spiders []*Spider) {
	for _, name := range []string{
		NameOfXici, NameOfKuai, NameOfYun,
		NameOfIphai, NameOfXila, NameOfNima,
		NameOfEightnine, NameOfHappy,
	} {
		spiders = append(spiders, NewSpider(name))
	}
	return
}

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
	case NameOfXila:
		return newSpider(name, xilaSpider{})
	case NameOfNima:
		return newSpider(name, nimaSpider{})
	case NameOfEightnine:
		return newSpider(name, eightnineSpider{})
	case NameOfHappy:
		return newSpider(name, happySpider{})
	default:
		return nil
	}
}

type xiciSpider struct{}

func (s xiciSpider) Urls() (urls []string) {
	baseURL := "https://www.xicidaili.com"
	for page := 1; page <= 5; page++ {
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

type xilaSpider struct{}

func (s xilaSpider) Urls() (urls []string) {
	return []string{"http://www.xiladaili.com/"}
}
func (s xilaSpider) Query() string {
	return `//*[@id="scroll"]/table/tbody/tr`
}
func (s xilaSpider) Parse(e *colly.XMLElement) (ip, port string) {
	if row := e.ChildText("td[1]"); strings.Count(row, ".") == 3 { // filter dirty data
		if result := strings.Split(row, ":"); len(result) == 2 {
			ip = result[0]
			port = result[1]
		}
	}
	return
}

type nimaSpider struct{}

func (s nimaSpider) Urls() (urls []string) {
	urls = append(urls, "http://www.nimadaili.com/")
	baseURL := "http://www.nimadaili.com"
	for page := 1; page <= 20; page++ {
		for _, domain := range []string{"gaoni", "http", "https"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", baseURL, domain, page))
		}
	}
	return
}
func (s nimaSpider) Query() string {
	return `//tbody/tr`
}
func (s nimaSpider) Parse(e *colly.XMLElement) (ip, port string) {
	row := e.ChildText("td[1]")
	if result := strings.Split(row, ":"); len(result) == 2 {
		ip = result[0]
		port = result[1]
	}
	return
}

type eightnineSpider struct{}

func (s eightnineSpider) Urls() (urls []string) {
	baseURL := "http://www.89ip.cn"
	for page := 1; page <= 15; page++ {
		urls = append(urls, fmt.Sprintf("%s/index_%d.html", baseURL, page))
	}
	return
}
func (s eightnineSpider) Query() string {
	return `//tbody/tr`
}
func (s eightnineSpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[1]")
	port = e.ChildText("td[2]")
	return
}

type happySpider struct{}

func (s happySpider) Urls() (urls []string) {
	baseURL := "http://ip.kxdaili.com"
	for page := 1; page <= 10; page++ {
		urls = append(urls, fmt.Sprintf("%s/ipList/%d.html#ip", baseURL, page))
	}
	return
}
func (s happySpider) Query() string {
	return `//*[@id="nav_btn01"]/div[5]/table/tbody/tr`
}
func (s happySpider) Parse(e *colly.XMLElement) (ip, port string) {
	ip = e.ChildText("td[1]")
	port = e.ChildText("td[2]")
	return
}
