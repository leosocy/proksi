package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

func main() {
	proxyChan := make(chan *proxy.Proxy)
	spider := spider.NewSpider(spider.NameOfXici)
	go spider.Do(proxyChan)
	for {
		select {
		case p := <-proxyChan:
			fmt.Printf("%+v\n", p)
		}
	}
}
