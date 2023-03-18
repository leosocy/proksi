// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/rs/zerolog"
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

type probeResult struct {
	prober    Prober
	protocols Protocols
	err       error
}

func newCombinedProber() *compositeProber {
	return &compositeProber{
		probers: []Prober{
			newSOCKS5Prober(),
			newSOCKS4Prober(),
			newHTTPSProber(),
			newHTTPProber(),
		},
		logger: zerolog.New(os.Stderr).With().Str("module", "protocol").Str("prober", "COMPOSITE").Logger(),
	}
}

func (p *compositeProber) Probe(ctx context.Context, addr string) (Protocols, error) {
	wg := sync.WaitGroup{}
	ch := make(chan probeResult)

	for _, prober := range p.probers {
		wg.Add(1)
		prober := prober
		go func() {
			defer wg.Done()
			protocol, err := prober.Probe(ctx, addr)
			ch <- probeResult{prober, protocol, err}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	protocols := NothingProtocols
	for pe := range ch {
		if pe.err == nil {
			protocols = protocols.Combine(pe.protocols)
		}
	}
	if protocols == NothingProtocols {
		p.logger.Info().Str("addr", addr).Msg("all probers failed, nothing protocols are supported")
	}

	return protocols, nil
}
