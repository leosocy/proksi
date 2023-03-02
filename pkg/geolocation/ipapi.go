// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package geolocation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/ratelimit"

	"github.com/leosocy/proksi/pkg/proxy"
)

type ipapiResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	CountryName string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	RegionCode  string  `json:"region"`
	City        string  `json:"city"`
	Lat         float32 `json:"lat"`
	Lon         float32 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
}

func (resp *ipapiResponse) isSuccess() bool {
	return resp.Status == "success"
}

func (resp *ipapiResponse) toGeolocation() *proxy.Geolocation {
	return &proxy.Geolocation{
		CountryName: resp.CountryName,
		CountryCode: resp.CountryCode,
		RegionName:  resp.RegionName,
		RegionCode:  resp.RegionCode,
		City:        resp.City,
		Lat:         resp.Lat,
		Lon:         resp.Lon,
		ISP:         resp.ISP,
		Org:         resp.Org,
	}
}

// IpapiGeolocator implements proxy.Geolocator, uses http://ip-api.com/json.
type IpapiGeolocator struct {
	baseURL string
	c       *http.Client
	limiter ratelimit.Limiter
}

func NewIpapiGeolocator() *IpapiGeolocator {
	loc := &IpapiGeolocator{
		baseURL: "http://ip-api.com/json/%s?lang=en&fields=status,country,countryCode,region,regionName,city,lat,lon,isp,org",
		c:       &http.Client{Timeout: 2 * time.Second},
		limiter: ratelimit.New(40, ratelimit.Per(time.Minute)),
	}
	return loc
}

func (loc *IpapiGeolocator) Name() proxy.GeolocatorName {
	return "ip-api"
}

func (loc *IpapiGeolocator) Locate(ctx context.Context, ip string) (*proxy.Geolocation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(loc.baseURL, ip), nil)
	if err != nil {
		return nil, err
	}

	resp, err := loc.c.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dto := ipapiResponse{}
	err = jsoniter.Unmarshal(b, &dto)
	if err != nil {
		return nil, err
	}
	if dto.isSuccess() {
		return nil, fmt.Errorf("geolocation: ip-api got error: %s when locate %s", dto.Message, ip)
	}

	return dto.toGeolocation(), nil
}
