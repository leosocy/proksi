// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"github.com/Sirupsen/logrus"
	"github.com/elazarl/goproxy"
	"net/http"
	"net/url"
)

type ProxyConn struct {

}

type PooledProxyConns struct {

}

// ProxyServer is a middleman between client and proxy server
// client <--> middleman <--> proxy server <--> destination
type ProxyServer struct {
	// 对goproxy.ProxyHttpServer封装，
}

func ListenAndServe() {
	middleman := goproxy.NewProxyHttpServer()
	middleman.Verbose = true
	proxy := "http://122.193.246.140:9999"
	// for request to http://xxx，middleman<->proxy server连接默认keepalive，
	// 好像默认75s，改天查一下。
	middleman.Tr.Proxy = func(request *http.Request) (*url.URL, error) {
		return url.Parse(proxy)
	}
	// for request to https://xxx，middleman<->proxy server连接每次关闭
	middleman.ConnectDial = middleman.NewConnectDialToProxy(proxy)
	//middleman.Tr.Dial = func(network, addr string) (c net.Conn, err error) {
	//	c, err = net.Dial(network, addr)
	//	if c, ok := c.(*net.TCPConn); err == nil && ok {
	//		err = c.SetKeepAlive(true)
	//	}
	//	return
	//}
	logrus.Fatal(http.ListenAndServe(":8081", middleman))
}
