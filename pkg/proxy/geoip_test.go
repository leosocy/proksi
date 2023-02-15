// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/leosocy/proksi/pkg/utils"
)

func fakeTooManyRequestsResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(""))
}

func fakeEmptyResp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func fakeIPAPISuccessResp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{
		"status": "success",
		"country": "China",
		"countryCode": "CN",
		"region": "JS",
		"regionName": "Jiangsu",
		"city": "Nanjing",
		"lat": 32.0617,
		"lon": 118.7778,
		"isp": "Chinanet"
	  }`))
}

func newMockedFetcher(name string, url string) (f GeoInfoFetcher) {
	switch name {
	case NameOfIPAPIFetcher:
		f = &ipAPIFetcher{
			baseFetcher: baseFetcher{tagName: "ip-api-json", baseURL: url},
			limiter:     &utils.RateLimiter{Delay: 10, Parallelism: 2},
		}
	}
	f.init()
	return
}

func TestGeoInfoFetcherDo(t *testing.T) {
	type args struct {
		ip          string
		fetcherName string
		fakeResp    http.HandlerFunc
	}
	tests := []struct {
		name     string
		args     args
		wantInfo *GeoInfo
		wantErr  bool
	}{
		{
			name:     "TooManyRequestsResponseStatus",
			args:     args{ip: "1.2.3.4", fetcherName: NameOfIPAPIFetcher, fakeResp: fakeTooManyRequestsResp},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name:     "EmptyResponse",
			args:     args{ip: "1.2.3.4", fetcherName: NameOfIPAPIFetcher, fakeResp: fakeEmptyResp},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "IPAPISuccessResponse",
			args: args{ip: "1.2.3.4", fetcherName: NameOfIPAPIFetcher, fakeResp: fakeIPAPISuccessResp},
			wantInfo: &GeoInfo{
				CountryName: "China",
				CountryCode: "CN",
				RegionName:  "Jiangsu",
				RegionCode:  "JS",
				City:        "Nanjing",
				Lat:         32.0617,
				Lon:         118.7778,
				ISP:         "Chinanet",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.args.fakeResp))
			defer ts.Close()
			f := newMockedFetcher(tt.args.fetcherName, ts.URL)
			gotInfo, err := f.Do(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeoInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("NewGeoInfo() = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func BenchmarkGeoInfoFetcherDo(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(fakeIPAPISuccessResp))
	defer ts.Close()
	f := newMockedFetcher(NameOfIPAPIFetcher, ts.URL)
	for i := 0; i < b.N; i++ {
		f.Do("1.2.3.4")
	}
}

func TestNewGeoInfoFetcher(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		args  args
		wantF bool
	}{
		{
			name:  "ValidName",
			args:  args{name: NameOfIPAPIFetcher},
			wantF: true,
		},
		{
			name:  "InvalidName",
			args:  args{name: "unknown"},
			wantF: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotF := NewGeoInfoFetcher(tt.args.name)
			if (gotF != nil) != tt.wantF {
				t.Errorf("NewGeoInfoFetcher() F = %v, wantF %v", gotF, tt.wantF)
				return
			}
		})
	}
}
