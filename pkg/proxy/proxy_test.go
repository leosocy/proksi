// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package proxy

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/Leosocy/gipp/mocks"
	"github.com/Leosocy/gipp/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestProxy_DetectAnonymity_Unknown(t *testing.T) {
	assert := assert.New(t)
	pxy := &Proxy{}
	g := new(mocks.RequestHeadersGetter)
	g.On("GetRequestHeaders", mock.Anything).Return(
		utils.HTTPRequestHeaders{}, errors.New("error occur"),
	)
	g.On("GetRequestHeadersUsingProxy", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "1.2.3.4"}, nil,
	)
	err := pxy.DetectAnonymity(g)
	g.AssertNumberOfCalls(t, "GetRequestHeadersUsingProxy", 0)
	assert.NotNil(err)
	assert.Equal(Unknown, pxy.Anon)
}

func TestProxy_DetectAnonymity_Transparent(t *testing.T) {
	assert := assert.New(t)
	pxy := &Proxy{}
	g := new(mocks.RequestHeadersGetter)
	g.On("GetRequestHeaders", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "1.2.3.4, 5.6.7.8"}, nil,
	)
	g.On("GetRequestHeadersUsingProxy", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "1.2.3.4"}, nil,
	)
	err := pxy.DetectAnonymity(g)
	g.AssertExpectations(t)
	assert.Nil(err)
	assert.Equal(Transparent, pxy.Anon)
}

func TestProxy_DetectAnonymity_Anonymous(t *testing.T) {
	assert := assert.New(t)
	pxy := &Proxy{}
	g := new(mocks.RequestHeadersGetter)
	g.On("GetRequestHeaders", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "1.2.3.4"}, nil,
	)
	g.On("GetRequestHeadersUsingProxy", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "5.6.7.8", Via: "squid"}, nil,
	)
	err := pxy.DetectAnonymity(g)
	g.AssertExpectations(t)
	assert.Nil(err)
	assert.Equal(Anonymous, pxy.Anon)
}

func TestProxy_DetectAnonymity_Elite(t *testing.T) {
	assert := assert.New(t)
	pxy := &Proxy{}
	g := new(mocks.RequestHeadersGetter)
	g.On("GetRequestHeaders", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "1.2.3.4"}, nil,
	)
	g.On("GetRequestHeadersUsingProxy", mock.Anything).Return(
		utils.HTTPRequestHeaders{XForwardedFor: "5.6.7.8"}, nil,
	)
	err := pxy.DetectAnonymity(g)
	g.AssertExpectations(t)
	assert.Nil(err)
	assert.Equal(Elite, pxy.Anon)
}
