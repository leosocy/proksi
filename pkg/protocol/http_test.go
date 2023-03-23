// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func serveFakeHTTPProxy(t *testing.T, ln net.Listener) {
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			t.Errorf("failed to read data: %v", err)
			return
		}

		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			t.Errorf("failed to write data: %v", err)
			return
		}
	}()
}

func TestHTTPProber_Probe(t *testing.T) {
	assert := assert.New(t)
	// create a listener to mock the proxy server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create proxy listener: %v", err)
	}
	serveFakeHTTPProxy(t, ln)

	prober := newHTTPProber()
	protocol, err := prober.Probe(context.Background(), ln.Addr().String())
	assert.Nil(err)
	assert.Equal(HTTP, protocol)

	// test probing a non-working proxy
	ln.Close()
	_, err = prober.Probe(context.Background(), ln.Addr().String())
	assert.NotNil(err)
}

var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICGDCCAb8CFEkSgqYhlT0+Yyr9anQNJgtclTL0MAoGCCqGSM49BAMDMIGOMQsw
CQYDVQQGEwJJTDEPMA0GA1UECAwGQ2VudGVyMQwwCgYDVQQHDANMb2QxEDAOBgNV
BAoMB0dvUHJveHkxEDAOBgNVBAsMB0dvUHJveHkxGjAYBgNVBAMMEWdvcHJveHku
Z2l0aHViLmlvMSAwHgYJKoZIhvcNAQkBFhFlbGF6YXJsQGdtYWlsLmNvbTAeFw0x
OTA1MDcxMTUwMThaFw0zOTA1MDIxMTUwMThaMIGOMQswCQYDVQQGEwJJTDEPMA0G
A1UECAwGQ2VudGVyMQwwCgYDVQQHDANMb2QxEDAOBgNVBAoMB0dvUHJveHkxEDAO
BgNVBAsMB0dvUHJveHkxGjAYBgNVBAMMEWdvcHJveHkuZ2l0aHViLmlvMSAwHgYJ
KoZIhvcNAQkBFhFlbGF6YXJsQGdtYWlsLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABDlH4YrdukPFAjbO8x+gR9F8ID7eCU8Orhba/MIblSRrRVedpj08lK+2
svyoAcrcDsynClO9aQtsC9ivZ+Pmr3MwCgYIKoZIzj0EAwMDRwAwRAIgGRSSJVSE
1b1KVU0+w+SRtnR5Wb7jkwnaDNxQ3c3FXoICIBJV/l1hFM7mbd68Oi5zLq/4ZsrL
98Bb3nddk2xys6a9
-----END CERTIFICATE-----`)

var localhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgEsc8m+2aZfagnesg
qMgXe8ph4LtVu2VOUYhHttuEDsChRANCAAQ5R+GK3bpDxQI2zvMfoEfRfCA+3glP
Dq4W2vzCG5Uka0VXnaY9PJSvtrL8qAHK3A7MpwpTvWkLbAvYr2fj5q9z
-----END PRIVATE KEY-----`)

func TestHTTPSProber_Probe(t *testing.T) {
	assert := assert.New(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create raw listener: %v", err)
	}
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Fatalf("failed to create cert: %v", err)
	}
	ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cert}})
	serveFakeHTTPProxy(t, ln)

	prober := newHTTPSProber()
	protocol, err := prober.Probe(context.Background(), ln.Addr().String())
	assert.Nil(err)
	assert.Equal(HTTPS, protocol)

	// test probing a non-working proxy
	ln.Close()
	_, err = prober.Probe(context.Background(), ln.Addr().String())
	assert.NotNil(err)
}
