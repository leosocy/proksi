package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
)

func main() {
	fetcher, _ := proxy.NewGeoInfoFetcher(proxy.NameOfIPAPIFetcher)
	info, _ := fetcher.Do("8.8.8.8")
	fmt.Printf("%+v", info)
}
