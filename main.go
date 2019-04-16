package main

import (
	"fmt"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
)

func main() {
	for {
		client := browser.Client{
			MaxPage: 5,
			Delay:   100 * time.Millisecond,
			Timeout: 5 * time.Second,
		}
		cache := browser.Cache{}
		b := browser.NewBrowser(client, cache)
		fmt.Println(b.Random())
		time.Sleep(time.Second)
		fmt.Println(time.Now())
	}
}
