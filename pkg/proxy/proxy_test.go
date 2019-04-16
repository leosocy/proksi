// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewProxy(t *testing.T) {
	type args struct {
		ip   string
		port string
	}
	tests := []struct {
		name    string
		args    args
		want    *Proxy
		wantErr bool
	}{
		{name: "t", args: args{ip: "115.196.59.38", port: "8118"}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(mockIPAPISuccessResp))
			defer ts.Close()
			fetchers = []fetcher{newMockFetcher(ipAPIFetcherName, ts.URL)}
			got, err := NewProxy(tt.args.ip, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}
