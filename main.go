package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

func main() {
	proxyChan := make(chan *proxy.Proxy)
	spiders := make([]*spider.Spider, 0, 16)
	for _, name := range []string{
		spider.NameOfIphai,
	} {
		spiders = append(spiders, spider.NewSpider(name))
	}
	for _, s := range spiders {
		go s.Crawl()
	}
	for {
		select {
		case p := <-proxyChan:
			fmt.Printf("%+v\n", p)
		}
	}
}
