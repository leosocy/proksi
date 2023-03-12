// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

type socks4Prober struct {
	dialer *net.Dialer
	logger zerolog.Logger
}

func newSOCKS4Prober() *socks4Prober {
	return &socks4Prober{
		dialer: &net.Dialer{},
		logger: zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "SOCKS4").Logger(),
	}
}

func lookupIPv4(host string) (net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}
		return ipv4, nil
	}
	return nil, errors.New(fmt.Sprintf("protocol: no IPv4 address found for host: %s", host))
}

func splitHostPort(addr string) (string, int, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	portnum, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	if 1 > portnum || portnum > 0xffff {
		return "", 0, errors.New("port number out of range " + port)
	}
	return host, portnum, nil
}

// +----+----+----+----+----+----+----+----+----+----+....+----+
// | VN | CD | DSTPORT |      DSTIP        | USERID       |NULL|
// +----+----+----+----+----+----+----+----+----+----+....+----+
// 1    1      2         4                  variable       1
// where:
//
// VN: SOCKS version number, should be 0x04 for SOCKS4
// CD: Command code, indicating the type of request. Possible values are:
//
//	0x01: Establish a TCP/IP stream connection
//	0x02: Establish a TCP/IP port binding
//	0x03: Upstream SOCKS server (unsupported by most SOCKS4 servers)
//
// DSTPORT: The port number of the target server, in network byte order (big endian)
// DSTIP: The IP address of the target server, in network byte order (4 bytes)
// USERID: The user ID string, variable length (maximum 255 bytes), terminated with a null byte (0x00)
func (p *socks4Prober) socks4Connect(conn net.Conn, addr string) ([]byte, error) {
	host, port, err := splitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip, err := lookupIPv4(host)
	if err != nil {
		return nil, err
	}
	req := make([]byte, 0, 9)
	req = append(req, 4, 1, byte(port>>8), byte(port))
	req = append(req, ip.To4()...)
	req = append(req, 0)

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}
	resp := make([]byte, 8)
	n, err := conn.Read(resp)
	return resp[:n], err
}

func (p *socks4Prober) parseConnectReply(resp []byte) error {
	if len(resp) != 8 {
		return errors.New("protocol: socks4 server does not respond properly")
	}
	code := resp[1]
	switch code {
	case 90:
		return nil
	case 91:
		return errors.New("protocol: socks4 connection request rejected or failed")
	case 92:
		return errors.New("protocol: socks4 connection request rejected because SOCKS server cannot connect to identd on the client")
	case 93:
		return errors.New("protocol: socks4 connection request rejected because the client program and identd report different user-ids")
	default:
		return errors.New("protocol: socks4 connection request failed, unknown code")
	}
}

func (p *socks4Prober) doProbe(ctx context.Context, addr string) error {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := p.socks4Connect(conn, "google.com:80")
	if err != nil {
		return err
	}
	err = p.parseConnectReply(resp)
	return nil
}

func (p *socks4Prober) Probe(ctx context.Context, addr string) (Protocols, error) {
	if err := p.doProbe(ctx, addr); err != nil {
		p.logger.Debug().Str("addr", addr).Err(err).Msg("")
		return NothingProtocols, err
	}
	p.logger.Debug().Str("addr", addr).Msg("success")
	return NewProtocols(SOCKS4), nil
}

type socks5Prober struct {
	dialer *net.Dialer
	logger zerolog.Logger
}

func newSOCKS5Prober() *socks5Prober {
	return &socks5Prober{
		dialer: &net.Dialer{},
		logger: zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "SOCKS5").Logger(),
	}
}

// +----+-----+-------+------+----------+----------+
// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
// +----+-----+-------+------+----------+----------+
// | 1  |  1  | X'00' |  1   | Variable |    2     |
// +----+-----+-------+------+----------+----------+
//
// Where:
//
//	o  VER    protocol version: X'05'
//	o  CMD
//	   o  CONNECT X'01'
//	   o  BIND X'02'
//	   o  UDP ASSOCIATE X'03'
//	o  RSV    RESERVED
//	o  ATYP   address type of following address
//	   o  IP V4 address: X'01'
//	   o  DOMAINNAME: X'03'
//	   o  IP V6 address: X'04'
//	o  DST.ADDR       desired destination address
//	o  DST.PORT desired destination port in network octet
//	   order
func (p *socks5Prober) socks5Connect(conn net.Conn, addr string) ([]byte, error) {
	req := []byte{
		0x05, // SOCKS version number, should be 0x05 for SOCKS5
		0x01, // the number of method identifier octets that appear in the METHODS field
		0x00, // NO AUTHENTICATION REQUIRED
	}
	_, err := conn.Write(req)
	if err != nil {
		return nil, err
	}
	resp := make([]byte, 4)
	_, err = conn.Read(resp)
	if err != nil {
		return nil, err
	}
	if resp[0] != 5 {
		return nil, errors.New("protocol: socks5 unexpected protocol version " + strconv.Itoa(int(resp[0])))
	}

	host, port, err := splitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip, err := lookupIPv4(host)
	if err != nil {
		return nil, err
	}
	req = req[:0]
	req = append(
		req,
		0x05,
		0x01, // CONNECT Command
		0x00, // RESERVED
		0x01, // IPV4 address
	)
	req = append(req, ip.To4()...)
	req = append(req, byte(port>>8), byte(port))
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}
	n, err := conn.Read(resp)
	return resp[:n], err
}

func (p *socks5Prober) parseConnectReply(resp []byte) error {
	if resp[0] != 5 {
		return errors.New("protocol: socks5 unexpected protocol version " + strconv.Itoa(int(resp[0])))
	}
	code := resp[1]
	switch code {
	case 0x00:
		return nil
	case 0x01:
		return errors.New("protocol: socks5 general SOCKS server failure")
	case 0x02:
		return errors.New("protocol: socks5 connection not allowed by ruleset")
	case 0x03:
		return errors.New("protocol: socks5 network unreachable")
	case 0x04:
		return errors.New("protocol: socks5 host unreachable")
	case 0x05:
		return errors.New("protocol: socks5 connection refused")
	case 0x06:
		return errors.New("protocol: socks5 TTL expired")
	case 0x07:
		return errors.New("protocol: socks5 command not supported")
	case 0x08:
		return errors.New("protocol: socks5 address type not supported")
	default:
		return errors.New("protocol: socks5 unknown code: " + strconv.Itoa(int(code)))
	}
}

func (p *socks5Prober) doProbe(ctx context.Context, addr string) error {
	conn, err := dialContext(p.dialer, ctx, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := p.socks5Connect(conn, "google.com:80")
	if err != nil {
		return err
	}
	err = p.parseConnectReply(resp)
	return err
}

func (p *socks5Prober) Probe(ctx context.Context, addr string) (Protocols, error) {
	if err := p.doProbe(ctx, addr); err != nil {
		p.logger.Debug().Str("addr", addr).Err(err).Msg("")
		return NothingProtocols, err
	}
	p.logger.Debug().Str("addr", addr).Msg("success")
	return NewProtocols(SOCKS5), nil
}
