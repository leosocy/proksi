// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"github.com/Sirupsen/logrus"
	"net"
	"sync"
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

type combinedProber struct {
	probers []Prober
}

type probeResult struct {
	prober    Prober
	protocols Protocols
	err       error
}

func newCombinedProber() *combinedProber {
	return &combinedProber{
		probers: []Prober{
			newSOCKS5Prober(),
			newSOCKS4Prober(),
			newHTTPSProber(),
			newHTTPProber(),
		},
	}
}

func (p *combinedProber) Probe(ctx context.Context, addr string) (Protocols, error) {
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
		if pe.err != nil {
			logrus.Debugf("protocol: %T err %v", pe.prober, pe.err)
		} else {
			logrus.Debugf("protocol: %T probe result: %s", pe.prober, pe.protocols)
			protocols = protocols.Combine(pe.protocols)
		}
	}
	if protocols == NothingProtocols {
		logrus.Warnf("protocol: all probes fail, nothing protocols are supported")
	}

	return protocols, nil
}
