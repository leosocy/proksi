// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"net"
)

func dialContext(dialer *net.Dialer, ctx context.Context, addr string) (net.Conn, error) {
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok && !deadline.IsZero() {
		err = conn.SetDeadline(deadline)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// ProbeProtocols
//
// It first, check for support SOCKS5. If it is supported, then SOCKS4 must not be supported,
// and it will return whether it supports http/https traffic.
// Otherwise, check for support SOCKS4. If it is supported, it will return whether is supports http/https traffic.
// Otherwise, check for support HTTPS.
// Finally, check for support HTTP.
func ProbeProtocols(ctx context.Context, addr string) (Protocols, error) {
	return EmptyProtocols, nil
}
