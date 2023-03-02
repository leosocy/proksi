// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"net"
)

// httpProber probe HTTP and HTTPS protocol.
type httpProber struct {
	dialer *net.Dialer
}

func newHTTPProber() *httpProber {
	return &httpProber{
		dialer: &net.Dialer{},
	}
}

func probeHTTPTraffic(conn net.Conn) error {
	_, err := conn.Write([]byte("GET http://httpbin.org/get HTTP/1.1\r\nHost: httpbin.org\r\n\r\n"))
	if err != nil {
		return err
	}
	resp := make([]byte, 1024)
	_, err = conn.Read(resp)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(resp, []byte("HTTP/1.1")) {
		return errors.New(fmt.Sprintf("protocol: http prober received invalid response %s", resp))
	}
	return nil
}

func (p *httpProber) Probe(ctx context.Context, addr string) (Protocols, error) {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return NothingProtocols, err
	}
	defer conn.Close()

	err = probeHTTPTraffic(conn)
	if err != nil {
		return NothingProtocols, err
	}
	return NewProtocols(HTTP), nil
}

type httpsProber struct {
	dialer          *net.Dialer
	TLSClientConfig *tls.Config
}

func newHTTPSProber() *httpsProber {
	return &httpsProber{
		dialer:          &net.Dialer{},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func (p *httpsProber) Probe(ctx context.Context, addr string) (Protocols, error) {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return NothingProtocols, err
	}
	defer conn.Close()

	conn = tls.Client(conn, p.TLSClientConfig)
	err = probeHTTPTraffic(conn)
	if err != nil {
		return NothingProtocols, err
	}

	return NewProtocols(HTTPS), nil
}
