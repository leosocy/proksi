// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testIP string = "1.2.3.4"

func mockTooManyRequestsResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(""))
}

func mockSuccessResp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{
		"status": "success",
		"country": "Canada",
		"countryCode": "CA",
		"region": "QC",
		"regionName": "Quebec",
		"city": "Montreal",
		"lat": 45.5808,
		"lon": -73.5825,
		"isp": "Le Groupe Videotron Ltee"
	  }`))
}

func TestNewGeoInfo(t *testing.T) {
	type args struct {
		ip       string
		mockFunc http.HandlerFunc
	}
	tests := []struct {
		name     string
		args     args
		wantInfo *GeoInfo
		wantErr  bool
	}{
		{
			name:     "TooManyRequestsResponseStatus",
			args:     args{ip: "1.2.3.4", mockFunc: mockTooManyRequestsResp},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name:     "SuccessResponse",
			args:     args{ip: "1.2.3.4", mockFunc: mockSuccessResp},
			wantInfo: &GeoInfo{CountryName: "Canada", RegionName: "Quebec", Lon: -73.5825},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ts := httptest.NewServer(http.HandlerFunc(tt.args.mockFunc))
			defer ts.Close()
			fetchers = []fetcher{&ipAPIFetcher{baseFetcher{tagName: "ip-api-json", baseURL: ts.URL}}}
			gotInfo, err := NewGeoInfo(tt.args.ip)
			assert.Equal(err != nil, tt.wantErr)
			if tt.wantInfo == nil {
				assert.Nil(gotInfo)
			} else {
				assert.Equal(tt.wantInfo.CountryName, gotInfo.CountryName)
				assert.Equal(tt.wantInfo.Lon, gotInfo.Lon)
			}
		})
	}
}

func BenchmarkNewGeoInfo(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(mockSuccessResp))
	defer ts.Close()
	fetchers = []fetcher{&ipAPIFetcher{baseFetcher{tagName: "ip-api-json", baseURL: ts.URL}}}
	for i := 0; i < b.N; i++ {
		NewGeoInfo(testIP)
	}
}
