package main

import (
	"fmt"

	ua "github.com/EDDYCJY/fake-useragent"
	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

func main() {

	proxies := make(chan *proxy.Proxy)
	urls := []string{"https://www.xicidaili.com/nt/"}
	s := spider.XiciSpider{BaseSpider: spider.BaseSpider{Name: "xici", UrlsFmt: urls}}
	s.Crawl(proxies)
	ua := ua.Random()
	fmt.Print(ua)
}
