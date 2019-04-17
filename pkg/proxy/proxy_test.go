// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert := assert.New(t)
			got, err := NewProxy(tt.args.ip, tt.args.port)
			assert.Equal(err != nil, tt.wantErr)
			if tt.want == nil {
				assert.Nil(got)
			} else {
				assert.Equal(tt.want.IP, got.IP)
				assert.Equal(tt.want.Port, got.Port)
			}
		})
	}
}
