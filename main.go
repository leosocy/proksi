package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
	"github.com/Leosocy/gipp/pkg/spider"
)

func main() {
	proxyChan := make(chan *proxy.Proxy)
	s := spider.NewSpiderName(spider.NameOfXici)
	s.Start()
	for {
		select {
		case p := <-proxyChan:
			fmt.Printf("%+v\n", p)
		}
	}
}
