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
	fakeHTTPBinIPToolXForwardedForBody string = `{
		"headers": {
		  "Host": "httpbin.org",
		  "Via": "1.1 squid",
		  "X-Forwarded-For": "1.2.3.4, 5.6.7.8", 
		  "X-Real-Ip": "9.10.11.12"
		}
	  }`
	fakeHTTPBinIPToolRealIPBody string = `{
		"headers": {
		  "Host": "httpbin.org",
		  "X-Real-Ip": "9.10.11.12"
		}
	  }`
	fakeHTTPBinIPToolEmptyBody string = `{}`
)

func TestHTTPBinIPTool_GetRequestHeaderUsingProxy(t *testing.T) {
	type args struct {
		proxyURL string
		fakeBody string
	}
	tests := []struct {
		name        string
		args        args
		wantHeaders HTTPRequestHeaders
		wantErr     bool
	}{
		{
			name: "WithXForwardedForBody",
			args: args{proxyURL: "", fakeBody: fakeHTTPBinIPToolXForwardedForBody},
			wantHeaders: HTTPRequestHeaders{
				XForwardedFor: "1.2.3.4, 5.6.7.8", XRealIP: "9.10.11.12", Via: "1.1 squid",
			},
			wantErr: false,
		},
		{
			name: "WithXRealIPBody",
			args: args{proxyURL: "", fakeBody: fakeHTTPBinIPToolRealIPBody},
			wantHeaders: HTTPRequestHeaders{
				XForwardedFor: "", XRealIP: "9.10.11.12", Via: "",
			},
			wantErr: false,
		},
		{
			name:        "WithEmptyBody",
			args:        args{proxyURL: "", fakeBody: fakeHTTPBinIPToolEmptyBody},
			wantHeaders: HTTPRequestHeaders{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.args.fakeBody))
			}))
			defer ts.Close()
			httpURLOfHTTPBin = ts.URL
			gotHeaders, err := HTTPBinIPTool{}.GetRequestHeaderUsingProxy(tt.args.proxyURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPBinIPTool.GetRequestHeaderUsingProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHeaders, tt.wantHeaders) {
				t.Errorf("HTTPBinIPTool.GetRequestHeaderUsingProxy() = %v, want %v", gotHeaders, tt.wantHeaders)
			}
		})
	}
}

func TestParsePublicIP(t *testing.T) {
	type args struct {
		headers HTTPRequestHeaders
	}
	tests := []struct {
		name    string
		args    args
		wantIP  net.IP
		wantErr bool
	}{
		{
			name: "XForwardedForExists",
			args: args{
				headers: HTTPRequestHeaders{
					XForwardedFor: "1.2.3.4, 5.6.7.8",
					XRealIP:       "9.10.11.12",
				},
			},
			wantIP: net.ParseIP("1.2.3.4"),
		},
		{
			name: "XForwardedForNotExists",
			args: args{
				headers: HTTPRequestHeaders{
					XRealIP: "9.10.11.12",
				},
			},
			wantIP: net.ParseIP("9.10.11.12"),
		},
		{
			name: "AllNotExists",
			args: args{
				headers: HTTPRequestHeaders{},
			},
			wantIP: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIP := tt.args.headers.ParsePublicIP()
			if !reflect.DeepEqual(gotIP, tt.wantIP) {
				t.Errorf("ParsePublicIP() = %v, want %v", gotIP, tt.wantIP)
			}
		})
	}
}

func BenchmarkHTTPBinIPTool_GetRequestHeaderUsingProxy(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fakeHTTPBinIPToolXForwardedForBody))
	}))
	defer ts.Close()
	httpURLOfHTTPBin = ts.URL
	for i := 0; i < b.N; i++ {
		HTTPBinIPTool{}.GetRequestHeaderUsingProxy("")
	}
}
