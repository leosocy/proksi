package main

import (
	"fmt"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/parnurzeal/gorequest"
)

func main() {
	ua := browser.Random()
	fmt.Println(ua)
	resp, body, errs := gorequest.New().Proxy("http://198.50.145.28:80").
		Timeout(100*time.Second).Get("http://httpbin.org/get?show_env=1").
		Set("User-Agent", ua).End()
	fmt.Println(resp, body, errs)
}
