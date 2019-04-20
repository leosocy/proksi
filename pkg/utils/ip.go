// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

// HTTPTool provides methods to get client information,
// such as public IP, useragent, and so on.
type HTTPTool interface {
}

var (
	httpURLOfHTTPBin  = "http://httpbin.org/get?show_env=1"
	httpsURLOfHTTPBin = "https://httpbin.org/get?show_env=1"
)

// HTTPBinIPTool provides methods to get client information,
// such as public IP, useragent, and so on.
// It gives the result by making a specific request to the
// `http(s)://httpbin.org` website and then parsing the response.
type HTTPBinIPTool struct {
	*gorequest.SuperAgent
}

// GetHTTPBinIPTool returns a tool instance.
func GetHTTPBinIPTool() HTTPBinIPTool {
	return HTTPBinIPTool{}
}

// GetPublicIP returns public ip where this program running.
func (t HTTPBinIPTool) GetPublicIP() (ip net.IP, err error) {
	var headers map[string]interface{}
	if headers, err = t.getResponseHeaders("", false); err == nil {
		return t.parsePublicIP(headers)
	}
	return
}

// GetPublicIPAndViaUsingProxy returns public ip and value of `Via` when using proxy.
func (t HTTPBinIPTool) GetPublicIPAndViaUsingProxy(proxyURL string) (ip net.IP, via string, err error) {
	var headers map[string]interface{}
	if headers, err = t.getResponseHeaders(proxyURL, false); err == nil {
		ip, err = t.parsePublicIP(headers)
		if via, found := headers["Via"]; found {
			return ip, via.(string), err
		}
	}
	return
}

// GetPublicIPUsingProxyAndHTTPS returns public ip where this program running
// when using proxy and make a `https` request to proxy server.
func (t HTTPBinIPTool) GetPublicIPUsingProxyAndHTTPS(proxyURL string) (ip net.IP, err error) {
	var headers map[string]interface{}
	if headers, err = t.getResponseHeaders(proxyURL, true); err == nil {
		return t.parsePublicIP(headers)
	}
	return
}

func (t HTTPBinIPTool) getResponseHeaders(proxyURL string, viaHTTPS bool) (headers map[string]interface{}, err error) {
	var reqURL string
	if viaHTTPS {
		reqURL = httpsURLOfHTTPBin
	} else {
		reqURL = httpURLOfHTTPBin
	}
	resp, body, errs := gorequest.New().
		Proxy(proxyURL).Timeout(100 * time.Second).
		Get(reqURL).EndBytes()
	if errs != nil || resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request %s failed. Proxy[%s]\tViaHTTPS[%t]", reqURL, proxyURL, viaHTTPS)
	}
	headers, err = t.unmarshalHeaders(body)
	return
}

func (t HTTPBinIPTool) unmarshalHeaders(body []byte) (headers map[string]interface{}, err error) {
	var bj map[string]interface{}
	if err = json.Unmarshal(body, &bj); err != nil {
		return nil, err
	}
	if headers, found := bj["headers"]; found {
		return headers.(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("`headers` not found in response body")
}

// parsePublicIP returns the public ip in headers.
// It will try parse from `X-Forwarded-For`,
// if failed, try parse from `X-Real-Ip`.
func (t HTTPBinIPTool) parsePublicIP(headers map[string]interface{}) (ip net.IP, err error) {
	if xForwardedFor, found := headers["X-Forwarded-For"]; found {
		for _, ipStr := range strings.Split(xForwardedFor.(string), ",") {
			if ip = net.ParseIP(strings.TrimSpace(ipStr)); ip != nil {
				return
			}
		}
	}
	if xRealIP, found := headers["X-Real-Ip"]; found {
		if ip = net.ParseIP(strings.TrimSpace(xRealIP.(string))); ip != nil {
			return
		}
	}
	return nil, fmt.Errorf("Can't parse public ip in headers")
}
