// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"net"
	"os"

	"github.com/rs/zerolog"
	"go.uber.org/multierr"
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

type compositeProber struct {
	probers []Prober
	logger  zerolog.Logger
}

func newCombinedProber() *compositeProber {
	return &compositeProber{
		probers: []Prober{
			newSOCKS4Prober(),
			newHTTPProber(),
			newSOCKS5Prober(),
			newHTTPSProber(),
		},
		logger: zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "composite").Logger(),
	}
}

func (p *compositeProber) Probe(ctx context.Context, addr string) (Protocol, error) {
	var me error
	for _, prober := range p.probers {
		protocol, err := prober.Probe(ctx, addr)
		if err != nil {
			me = multierr.Append(me, err)
		} else {
			return protocol, nil
		}
	}
	return Nothing, me
}

var (
	prober = newCombinedProber()
)

// Probe uses the default composite prober to probe the given address using multiple Prober.
// It returns the first successful protocol and any errors encountered during probing if nothing protocol supported.
func Probe(ctx context.Context, addr string) (Protocol, error) {
	return prober.Probe(ctx, addr)
}
