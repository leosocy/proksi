// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package backend

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/leosocy/proksi/pkg/proxy"
	"github.com/leosocy/proksi/pkg/storage"
)

type BackendTestSuite struct {
	suite.Suite
	backends []Backend
}

func (suite *BackendTestSuite) SetupTest() {
	suite.backends = []Backend{
		NewInMemoryBackend(),
	}
	// insert and assert some proxies
	for _, s := range suite.backends {
		// insert invalid proxy
		err := s.Insert(nil)
		suite.Equal(err, ErrProxyInvalid)
		// insert two proxy
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("1.2.3.4"), Port: 80, Score: 50})
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 80})
		suite.Nil(err)
		suite.Equal(uint(2), s.Len())
		// insert another proxy
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("9.10.11.12"), Port: 80, Score: 30})
		suite.Equal(uint(3), s.Len())
		// insert proxy with same IP
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("9.10.11.12"), Port: 80, Score: 40})
		suite.Equal(err, ErrProxyDuplicated)
		suite.Equal(uint(3), s.Len())
	}
}

func (suite *BackendTestSuite) TestSelect() {
	for _, s := range suite.backends {
		// no options
		pxys, err := s.Select()
		suite.Equal(s.Len(), uint(len(pxys)))
		suite.Nil(err)
		// with limit
		pxys, err = s.Select(storage.WithLimit(1))
		suite.Equal(1, len(pxys))
		suite.Nil(err)
		// with offset
		pxys, err = s.Select(storage.WithOffset(1))
		suite.Equal(2, len(pxys))
		suite.Nil(err)
		// filter score
		pxys, err = s.Select(storage.WithFilter(storage.FilterUptime(60)))
		suite.Equal(1, len(pxys))
		suite.True(pxys[0].Score >= 60)
		// filter score none available
		pxys, err = s.Select(storage.WithFilter(storage.FilterUptime(100)))
		suite.NotNil(err)
		// filter and offset, limit
		pxys, err = s.Select(storage.WithFilter(storage.FilterUptime(50)), storage.WithLimit(10))
		suite.Equal(2, len(pxys))
		// filter and offset out of range
		pxys, err = s.Select(storage.WithFilter(storage.FilterUptime(50)), storage.WithOffset(10))
		suite.NotNil(err)
	}
}

func (suite *BackendTestSuite) TestSearch() {
	for _, s := range suite.backends {
		pxy := s.Search(net.ParseIP("5.6.7.8"))
		suite.Equal(pxy.IP.String(), "5.6.7.8")
		// not found
		pxy = s.Search(net.ParseIP("8.8.8.8"))
		suite.Nil(pxy)
	}
}

func (suite *BackendTestSuite) TestDelete() {
	for _, s := range suite.backends {
		// does not exists
		err := s.Delete(&proxy.Proxy{IP: net.ParseIP("8.8.8.8")})
		suite.Equal(err, ErrProxyDoesNotExists)
		// normal
		bLen := s.Len()
		err = s.Delete(&proxy.Proxy{IP: net.ParseIP("5.6.7.8")})
		searchP := s.Search(net.ParseIP("5.6.7.8"))
		suite.Nil(err)
		suite.Nil(searchP)
		suite.Equal(bLen-1, s.Len())
	}
}

func (suite *BackendTestSuite) TestTopK() {
	for _, s := range suite.backends {
		bps := s.TopK(2)
		suite.Equal(2, len(bps))
		suite.True(bps[0].Score > bps[1].Score)
		suite.Equal(3, len(s.TopK(0)))
	}
}

func (suite *BackendTestSuite) TestUpdate() {
	for _, s := range suite.backends {
		// does not exists
		err := s.Update(&proxy.Proxy{IP: net.ParseIP("6.7.8.9"), Port: 80, Score: 50})
		suite.Equal(err, ErrProxyDoesNotExists)
		// normal
		p := &proxy.Proxy{IP: net.ParseIP("1.2.3.4"), Port: 80, Score: 50}
		p.Score = 90
		err = s.Update(p)
		bp := s.TopK(1)[0]
		suite.Nil(err)
		suite.Equal(p.IP, bp.IP)
	}
}

func (suite *BackendTestSuite) TestInsertOrUpdate() {
	for _, s := range suite.backends {
		p := &proxy.Proxy{IP: net.ParseIP("6.6.6.6"), Port: 80, Score: 50}
		inserted, err := s.InsertOrUpdate(p)
		suite.Nil(err)
		suite.True(inserted)
		// update
		p.Score = 100
		inserted, err = s.InsertOrUpdate(p)
		suite.Nil(err)
		suite.False(inserted)
		sp := s.Search(p.IP)
		suite.Equal(int8(100), sp.Score)
	}
}

func (suite *BackendTestSuite) TestIter() {
	for _, s := range suite.backends {
		total := 0
		s.Iter(func(pxy *proxy.Proxy) bool {
			total++
			if total >= 2 {
				return false
			}
			return true
		})
		suite.Equal(2, total)
	}
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}
