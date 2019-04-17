// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// ParsePublicIPFromResponseBody returns the public ip
// of client which make the request.
// It will try parse from `X-Forwarded-For`,
// if failed, try parse from `X-Real-Ip`.
func ParsePublicIPFromResponseBody(body []byte) (ip net.IP, err error) {
	var bj map[string]interface{}
	if err = json.Unmarshal(body, &bj); err != nil {
		return nil, err
	}
	if xForwardedFor, ok := bj["X-Forwarded-For"]; ok {
		for _, ipStr := range strings.Split(xForwardedFor.(string), ",") {
			if ip = net.ParseIP(strings.TrimSpace(ipStr)); ip != nil {
				return
			}
		}
	}
	if xRealIP, ok := bj["X-Real-Ip"]; ok {
		if ip = net.ParseIP(strings.TrimSpace(xRealIP.(string))); ip != nil {
			return
		}
	}
	return nil, fmt.Errorf("can't parse public ip")
}
