// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder *Builder
		want    *Proxy
		wantErr bool
	}{
		{
			name:    "IPPortStringWithSpace",
			builder: NewBuilder().IP("1.2.3.4 ").Port(" 1234"),
			want:    &Proxy{AddrPort: netip.MustParseAddrPort("1.2.3.4:1234")},
			wantErr: false,
		},
		{
			name:    "InvalidIP",
			builder: NewBuilder().AddrPort("1.2.3.:1234"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "InvalidPort",
			builder: NewBuilder().AddrPort("1.2.3.4:"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "InvalidIPPort",
			builder: NewBuilder().IP("1.2.3.").Port(""),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.builder.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder error got = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if (tt.want == nil && got != nil) || (tt.want != nil && got == nil) || (tt.want != nil && !reflect.DeepEqual(got.AddrPort, tt.want.AddrPort)) {
				t.Errorf("Builder proxy got = %+v, want %+v", got, tt.want)
				return
			}
		})
	}
}

func BenchmarkNewBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewBuilder().AddrPort("127.0.0.1:8080").Build()
	}
}

func TestProxy_Equal(t *testing.T) {
	assert := assert.New(t)
	one, _ := NewBuilder().AddrPort("1.2.3.4:80").Build()
	another, _ := NewBuilder().AddrPort("1.2.3.4:8080").Build()
	assert.False(one.Equal(another))
}
