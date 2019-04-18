// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import (
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	httpBinIPToolXForwardedForBody string = `{
		"headers": {
		  "Host": "httpbin.org",
		  "Via": "1.1 squid",
		  "X-Forwarded-For": "1.2.3.4, 5.6.7.8", 
		  "X-Real-Ip": "9.10.11.12"
		}
	  }`
	httpBinIPToolRealIPBody string = `{
		"headers": {
		  "Host": "httpbin.org",
		  "X-Real-Ip": "9.10.11.12"
		}
	  }`
	httpBinIPToolEmptyBody string = `{}`
)

func TestHTTPBinIPTool_GetPublicIPAndViaUsingProxy(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name    string
		args    args
		wantIP  net.IP
		wantVia string
		wantErr bool
	}{
		{
			name:    "GetPublicIPAndViaUsingProxyWithXForwardedForBody",
			args:    args{body: httpBinIPToolXForwardedForBody},
			wantIP:  net.ParseIP("1.2.3.4"),
			wantVia: "1.1 squid",
			wantErr: false,
		},
		{
			name:    "GetPublicIPAndViaUsingProxyWithRealIPBody",
			args:    args{body: httpBinIPToolRealIPBody},
			wantIP:  net.ParseIP("9.10.11.12"),
			wantVia: "",
			wantErr: false,
		},
		{
			name:    "GetPublicIPAndViaUsingProxyWithEmptyBody",
			args:    args{body: httpBinIPToolEmptyBody},
			wantIP:  nil,
			wantVia: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.args.body))
			}))
			defer ts.Close()
			httpURLOfHTTPBin = ts.URL
			httpsURLOfHTTPBin = ts.URL

			gotIP, gotVia, err := GetHTTPBinIPTool().GetPublicIPAndViaUsingProxy("")
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPBinIPTool.GetPublicIPAndViaUsingProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIP, tt.wantIP) {
				t.Errorf("HTTPBinIPTool.GetPublicIPAndViaUsingProxy() gotIp = %v, want %v", gotIP, tt.wantIP)
			}
			if gotVia != tt.wantVia {
				t.Errorf("HTTPBinIPTool.GetPublicIPAndViaUsingProxy() gotVia = %v, want %v", gotVia, tt.wantVia)
			}
		})
	}
}

func BenchmarkHTTPBinIPTool_GetPublicIPAndViaUsingProxy(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(httpBinIPToolRealIPBody))
	}))
	defer ts.Close()
	httpURLOfHTTPBin = ts.URL
	httpsURLOfHTTPBin = ts.URL
	for i := 0; i < b.N; i++ {
		GetHTTPBinIPTool().GetPublicIPAndViaUsingProxy("")
	}
}
