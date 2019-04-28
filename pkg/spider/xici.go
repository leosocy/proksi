// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package spider

import (
	"fmt"
)

const xiciBaseURL = "https://www.xicidaili.com"

func buildXiciUrls() (urls []string) {
	for page := 1; page <= 2; page++ {
		for _, domain := range []string{"nn", "nt", "wn", "wt"} {
			urls = append(urls, fmt.Sprintf("%s/%s/%d", xiciBaseURL, domain, page))
		}
	}
	return
}
