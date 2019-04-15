// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/parnurzeal/gorequest"
)

// GeoInfo 包括ip的地理位置相关信息，包括国家省市，运营商，经纬度等等
// ip-api-json tag用于从 `http://www.ip-api.com/docs/api:json` 中拉取信息
type GeoInfo struct {
	CountryName string  `ip-api-json:"country"`     // e.g. China
	CountryCode string  `ip-api-json:"countryCode"` // e.g. CN
	RegionName  string  `ip-api-json:"regionName"`  // e.g. Jiangsu
	RegionCode  string  `ip-api-json:"region"`      // e.g. JS
	City        string  `ip-api-json:"city"`        // e.g. Nanjing
	Lat         float32 `ip-api-json:"lat"`         // e.g. 32.0617
	Lon         float32 `ip-api-json:"lon"`         // e.g. 118.7778
	ISP         string  `ip-api-json:"isp"`         // e.g. Chinanet
}

var request *gorequest.SuperAgent
var fetchers []fetcher

func init() {
	request = gorequest.New()
	fetchers = append(fetchers,
		&ipAPIFetcher{baseFetcher{tagName: "ip-api-json", baseURL: "http://ip-api.com"}})
}

// NewGeoInfo returns the geo information for ip.
// It will get a fetcher from the `fetchers`,
// make a request, and parse it
// until the parse succeeds or the loop ends
func NewGeoInfo(ip string) (info *GeoInfo, err error) {
	for _, f := range fetchers {
		body := []byte{}
		if body, err = f.fetch(ip); err == nil {
			if info, err := f.unmarshal(body); err == nil {
				return info, nil
			}
		}
	}
	return
}

type fetcher interface {
	init()
	fetch(ip string) (body []byte, err error)
	unmarshal(body []byte) (info *GeoInfo, err error)
}

type baseFetcher struct {
	tagName      string
	baseURL      string
	urlFormatter string

	sync.Once
}

func (f *baseFetcher) unmarshal(body []byte) (info *GeoInfo, err error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	info = &GeoInfo{}
	rv := reflect.ValueOf(info).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		tagValue := rt.Field(i).Tag.Get(f.tagName)
		if tagValue == "" || tagValue == "-" {
			continue
		}
		if v, found := data[tagValue]; found {
			rv.Field(i).Set(reflect.ValueOf(v).Convert(rt.Field(i).Type))
		}
	}
	return info, nil
}

// ipAPIFetcher see document: `http://www.ip-api.com/docs/api:json`
type ipAPIFetcher struct {
	baseFetcher
}

func (f *ipAPIFetcher) init() {
	var sb strings.Builder
	sb.Grow(128)
	sb.WriteString(f.baseURL)
	sb.WriteString("/json/%s?fields=status,message")
	t := reflect.TypeOf(GeoInfo{})
	for i := 0; i < t.NumField(); i++ {
		tagValue := t.Field(i).Tag.Get(f.tagName)
		if tagValue == "" || tagValue == "-" {
			continue
		}
		sb.WriteString(fmt.Sprintf(",%s", tagValue))
	}
	f.urlFormatter = sb.String()
}

func (f *ipAPIFetcher) fetch(ip string) (body []byte, err error) {
	f.Once.Do(f.init)
	url := fmt.Sprintf(f.urlFormatter, ip)
	resp, body, errs := request.Get(url).EndBytes()
	if resp.StatusCode != http.StatusOK || errs != nil {
		return nil, fmt.Errorf("fetch info from %s failed", url)
	}
	return body, nil
}
