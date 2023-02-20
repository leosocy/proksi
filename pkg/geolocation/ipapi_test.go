// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package geolocation

import (
	"context"
	"fmt"
	"github.com/leosocy/proksi/pkg/proxy"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func fakeTooManyRequestsResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(""))
}

func fakeEmptyResp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func fakeIpapiResp(status, message string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
		"status": "%s",
		"message": "%s",
		"country": "United States",
		"countryCode": "US",
		"region": "VA",
		"regionName": "Virginia",
		"city": "Ashburn",
		"lat": 39.03,
		"lon": -77.5,
		"isp": "Google LLC",
		"org": "Google Public DNS"
	  }`, status, message)))
	}
}

func TestIpapiGeoLocator(t *testing.T) {
	type args struct {
		ip          string
		fetcherName string
		fakeResp    http.HandlerFunc
	}
	tests := []struct {
		name     string
		args     args
		wantInfo *proxy.Geolocation
		wantErr  bool
	}{
		{
			name:     "TooManyRequestsResponseStatus",
			args:     args{ip: "1.2.3.4", fakeResp: fakeTooManyRequestsResp},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name:     "EmptyResponse",
			args:     args{ip: "1.2.3.4", fakeResp: fakeEmptyResp},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name:     "IpapiFailResponse",
			args:     args{ip: "1.2.3.4", fakeResp: fakeIpapiResp("fail", "invalid query")},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "IpapiSuccessResponse",
			args: args{ip: "1.2.3.4", fakeResp: fakeIpapiResp("success", "")},
			wantInfo: &proxy.Geolocation{
				CountryName: "United States",
				CountryCode: "US",
				RegionName:  "Virginia",
				RegionCode:  "VA",
				City:        "Ashburn",
				Lat:         39.03,
				Lon:         -77.5,
				ISP:         "Google LLC",
				Org:         "Google Public DNS",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.args.fakeResp)
			defer ts.Close()
			locator := NewIpapiGeolocator()
			locator.baseURL = fmt.Sprintf("%s%s", ts.URL, "/json/%s?lang=en&fields=status,country,countryCode,region,regionName,city,lat,lon,isp,org")
			gotInfo, err := locator.Locate(context.Background(), tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("got = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func BenchmarkIpapiGeoLocator(b *testing.B) {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(fakeIpapiResp("success", "")))
	defer ts.Close()
	locator := NewIpapiGeolocator()
	locator.baseURL = fmt.Sprintf("%s%s", ts.URL, "/json/%s?lang=en&fields=status,country,countryCode,region,regionName,city,lat,lon,isp,org")

	for i := 0; i < b.N; i++ {
		locator.Locate(ctx, "1.2.3.4")
	}
}
