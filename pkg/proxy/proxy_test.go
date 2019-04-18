// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net"
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
		{
			name:    "IPPortStringWithSpace",
			args:    args{ip: "1.2.3.4 ", port: "1234"},
			want:    &Proxy{IP: net.ParseIP("1.2.3.4"), Port: 1234},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProxy(tt.args.ip, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.IP, tt.want.IP) {
				t.Errorf("NewProxy().IP = %v, want %v", got.IP, tt.want)
			}
		})
	}
}
