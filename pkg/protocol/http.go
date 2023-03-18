// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// httpProber probe HTTP and HTTPS protocol.
type httpProber struct {
	dialer *net.Dialer
	logger zerolog.Logger
}

func newHTTPProber() *httpProber {
	return &httpProber{
		dialer: &net.Dialer{},
		logger: zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "HTTP").Logger(),
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

func (p *httpProber) doProber(ctx context.Context, addr string) error {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = probeHTTPTraffic(conn)
	return err
}

func (p *httpProber) Probe(ctx context.Context, addr string) (Protocols, error) {
	if err := p.doProber(ctx, addr); err != nil {
		p.logger.Debug().Err(err).Str("addr", addr).Msg("")
		return NothingProtocols, err
	}
	p.logger.Debug().Str("addr", addr).Msg("success")
	return NewProtocols(HTTP), nil
}

type httpsProber struct {
	TLSClientConfig *tls.Config
	dialer          *net.Dialer
	logger          zerolog.Logger
}

func newHTTPSProber() *httpsProber {
	return &httpsProber{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		dialer:          &net.Dialer{},
		logger:          zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "HTTPS").Logger(),
	}
}

func (p *httpsProber) doProbe(ctx context.Context, addr string) error {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn = tls.Client(conn, p.TLSClientConfig)
	err = probeHTTPTraffic(conn)
	return err
}

func (p *httpsProber) Probe(ctx context.Context, addr string) (Protocols, error) {
	if err := p.doProbe(ctx, addr); err != nil {
		p.logger.Debug().Err(err).Str("addr", addr).Msg("")
		return NothingProtocols, err
	}
	p.logger.Debug().Str("addr", addr).Msg("success")
	return NewProtocols(HTTPS), nil
}
