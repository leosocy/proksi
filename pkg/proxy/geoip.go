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

	"github.com/parnurzeal/gorequest"
)

// GeoInfo 包括ip的地理位置相关信息，包括国家省市，运营商，经纬度等等
// ip-api-json tag用于ip-api fetcher从 `http://www.ip-api.com/docs/api:json` 中拉取信息
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

const (
	// NameOfIPAPIFetcher name of ip-api fetcher implements.
	NameOfIPAPIFetcher string = "ip-api"
)

// GeoInfoFetcher have some functions which use to fetch geo information from specify url.
// init(): you can init url formatter or some other initial work.
// fetch(): you can request to url with ip.
// unmarshal(): unmarshal to GeoInfo struct accord to the tag name which define in the struct.
type GeoInfoFetcher interface {
	// Do returns the geo information for ip.
	// Use the fetcher to make a fetching request,
	// and parse response body.
	Do(ip string) (info *GeoInfo, err error)
	init()
	fetch(ip string) (body []byte, err error)
	unmarshal(body []byte) (info *GeoInfo, err error)
}

// NewGeoInfoFetcher returns a fetcher for name.
func NewGeoInfoFetcher(name string) (f GeoInfoFetcher, err error) {
	switch name {
	case NameOfIPAPIFetcher:
		f = &ipAPIFetcher{
			baseFetcher{tagName: "ip-api-json", baseURL: "http://ip-api.com"},
		}
	default:
		return f, fmt.Errorf("geo info fetcher name not support")
	}
	f.init()
	return
}

type baseFetcher struct {
	tagName      string
	baseURL      string
	urlFormatter string
}

func (f *baseFetcher) Do(ip string) (info *GeoInfo, err error) {
	var body []byte
	if body, err = f.fetch(ip); err == nil {
		info, err = f.unmarshal(body)
	}
	return
}

func (f *baseFetcher) fetch(ip string) (body []byte, err error) {
	url := fmt.Sprintf(f.urlFormatter, ip)
	resp, body, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil || resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch info from %s failed", url)
	}
	return body, nil
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
