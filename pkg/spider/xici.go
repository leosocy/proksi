// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Leosocy/gipp/pkg/proxy"
)

// XiciSpider root url: https://www.xicidaili.com/nn/
type XiciSpider struct {
	BaseSpider
}

func (s *XiciSpider) Crawl(chan<- *proxy.Proxy) {

	req, err := http.NewRequest("GET", "http://www.xicidaili.com/nn/", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("User-Agent", "123")
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Print(string(body))

}
