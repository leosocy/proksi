// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package spider

import (
	"github.com/gocolly/colly/v2"
	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestNewRegexParser(t *testing.T) {
	type args struct {
		name   string
		config *RegexpParserConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *regexParser
		wantErr bool
	}{
		{
			name:    "InvalidRegexp",
			args:    args{"test", &RegexpParserConfig{Expr: struct{ ipport string }{ipport: "\\"}}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "InvalidRegexpSubexp",
			args:    args{"test", &RegexpParserConfig{Expr: struct{ ipport string }{ipport: `(?\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(?P<port>\d+`}}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Ok",
			args: args{"test", &RegexpParserConfig{Expr: struct{ ipport string }{ipport: `(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(?P<port>\d{1,5})`}}},
			want: &regexParser{
				ipport: struct {
					regex   *regexp.Regexp
					ipIdx   int
					portIdx int
				}{regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(?P<port>\d{1,5})`), 1, 2},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRegexParser(tt.args.name, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRegexParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil && tt.want != nil) || (got != nil && tt.want == nil) || (got != nil && !reflect.DeepEqual(got.ipport, tt.want.ipport)) {
				t.Errorf("newRegexParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testCollector struct {
	proxies []*proxy.Proxy
}

func (c *testCollector) Collect(ps ...*proxy.Proxy) {
	c.proxies = append(c.proxies, ps...)
}

func (c *testCollector) Close() error {
	return nil
}

func flatten(proxies []*proxy.Proxy) string {
	elems := make([]string, 0, len(proxies))
	for _, pxy := range proxies {
		elems = append(elems, pxy.String())
	}
	return strings.Join(elems, ",")
}

func TestRegexParser_HandleResponse(t *testing.T) {
	parser, _ := newRegexParser("test", &RegexpParserConfig{Expr: struct{ ipport string }{ipport: `(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(?P<port>\d{1,5})`}})

	type args struct {
		resp *colly.Response
	}
	tests := []struct {
		name        string
		args        args
		wantProxies []*proxy.Proxy
	}{
		{
			name: "ValidResponse",
			args: args{&colly.Response{
				StatusCode: 200,
				Body:       []byte("  192.168.0.1:8080\n  127.0.0.1:3128"),
			}},
			wantProxies: []*proxy.Proxy{
				proxy.NewBuilder().AddrPort("192.168.0.1:8080").MustBuild(),
				proxy.NewBuilder().AddrPort("127.0.0.1:3128").MustBuild(),
			},
		},
		{
			name: "InvalidResponse",
			args: args{&colly.Response{
				StatusCode: 200,
				Body:       []byte("invalid response"),
			}},
			wantProxies: []*proxy.Proxy{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &testCollector{}
			parser.HandleResponse(tt.args.resp, c)
			assert.Equal(t, flatten(tt.wantProxies), flatten(c.proxies))
		})
	}
}
