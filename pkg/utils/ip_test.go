// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package utils

import (
	"net"
	"reflect"
	"testing"
)

const (
	xForwardedForBody string = `{
		"X-Forwarded-For": "1.2.3.4, 5.6.7.8", 
		"X-Real-Ip": "9.10.11.12"
	  }`
	xRealIPBody string = `{
		"X-Real-Ip": "9.10.11.12"
	  }`
	nothingBody string = `{}`
)

func TestParsePublicIPFromResponseBody(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		wantIP  net.IP
		wantErr bool
	}{
		{
			name:    "HasXForwardedFor",
			args:    args{body: []byte(xForwardedForBody)},
			wantIP:  net.ParseIP("1.2.3.4"),
			wantErr: false,
		},
		{
			name:    "HasXRealIP",
			args:    args{body: []byte(xRealIPBody)},
			wantIP:  net.ParseIP("9.10.11.12"),
			wantErr: false,
		},
		{
			name:    "HasNothingIP",
			args:    args{body: []byte(nothingBody)},
			wantIP:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIP, err := ParsePublicIPFromResponseBody(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicIPFromResponseBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIP, tt.wantIP) {
				t.Errorf("ParsePublicIPFromResponseBody() = %v, want %v", gotIP, tt.wantIP)
			}
		})
	}
}
