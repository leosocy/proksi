package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

func main() {
	proxyChan := make(chan *proxy.Proxy)
	s := spider.NewSpiderName(spider.NameOfKuai)
	s.Crawl()
	for {
		select {
		case p := <-proxyChan:
			fmt.Printf("%+v\n", p)
		}
	}
}
