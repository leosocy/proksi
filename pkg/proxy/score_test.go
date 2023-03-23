// Copyright (c) 2019 leosocy, leosocy@gmail.com

// Use of this source code is governed by a MIT-style license

// that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBatchHTTPSScorer(t *testing.T) {
	defer func() {
		r := recover()
		r, ok := r.(error)
		assert.True(t, ok)
	}()
	s := NewBatchHTTPSScorer([]string{"www.test.com", "www.test.com"})
	assert.Equal(t, 100*time.Second, s.(*BatchHTTPSScorer).timeout)
	// test panic
	NewBatchHTTPSScorer([]string{"www.test.com"})
}

func TestBatchHTTPSScorerCalculate(t *testing.T) {
	pxy, _ := NewBuilder().AddrPort("1.2.3.4:80").Build()
	pxy.IP = []byte("")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	s := NewBatchHTTPSScorer([]string{ts.URL, ts.URL})
	s.Score(pxy)
	assert.EqualValues(t, 0, pxy.Score)
}
