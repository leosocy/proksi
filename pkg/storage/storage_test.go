// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Leosocy/IntelliProxy/pkg/proxy"
)

type StorageTestSuite struct {
	suite.Suite
	storages []Storage
}

func (suite *StorageTestSuite) SetupTest() {
	suite.storages = []Storage{
		NewInMemoryStorage(),
	}
}

func (suite *StorageTestSuite) TestInsert() {
	for _, s := range suite.storages {
		// insert invalid proxy
		err := s.Insert(nil)
		suite.Equal(err, ErrProxyInvalid)
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("1.2.3.4"), Port: 80, Score: 0})
		suite.Equal(err, ErrProxyInvalid)
		// insert one proxy
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 100})
		suite.Nil(err)
		suite.Equal(uint(1), s.Len())
		// insert another proxy
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("9.10.11.12"), Port: 80, Score: 100})
		suite.Equal(uint(2), s.Len())
		// insert proxy with same IP
		err = s.Insert(&proxy.Proxy{IP: net.ParseIP("9.10.11.12"), Port: 80, Score: 50})
		suite.Equal(err, ErrProxyDuplicated)
		suite.Equal(uint(2), s.Len())
	}
}

func (suite *StorageTestSuite) TestSearch() {
	for _, s := range suite.storages {
		s.Insert(&proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 100})
		pxy := s.Search(net.ParseIP("5.6.7.8"))
		suite.Equal(pxy.IP.String(), "5.6.7.8")
		// not found
		pxy = s.Search(net.ParseIP("8.8.8.8"))
		suite.Nil(pxy)
	}
}

func (suite *StorageTestSuite) TestDelete() {
	for _, s := range suite.storages {
		p := &proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 100}
		s.Insert(p)
		// does not exists
		err := s.Delete(net.ParseIP("8.8.8.8"))
		suite.Equal(err, ErrProxyDoesNotExists)
		// normal
		err = s.Delete(p.IP)
		searchP := s.Search(p.IP)
		suite.Nil(err)
		suite.Nil(searchP)
		suite.Equal(uint(0), s.Len())
	}
}

func (suite *StorageTestSuite) TestBest() {
	for _, s := range suite.storages {
		// empty
		bp := s.Best()
		suite.Nil(bp)
		// normal
		p1 := &proxy.Proxy{IP: net.ParseIP("1.2.3.4"), Port: 80, Score: 50}
		p2 := &proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 80}
		s.Insert(p1)
		s.Insert(p2)
		bp = s.Best()
		suite.Equal(p2.IP, bp.IP)
	}
}

func (suite *StorageTestSuite) TestUpdate() {
	for _, s := range suite.storages {
		p1 := &proxy.Proxy{IP: net.ParseIP("1.2.3.4"), Port: 80, Score: 50}
		p2 := &proxy.Proxy{IP: net.ParseIP("5.6.7.8"), Port: 80, Score: 80}
		s.Insert(p1)
		s.Insert(p2)
		// does not exists
		err := s.Update(&proxy.Proxy{IP: net.ParseIP("6.7.8.9"), Port: 80, Score: 50})
		suite.Equal(err, ErrProxyDoesNotExists)
		// normal
		p1.Score = 90
		err = s.Update(p1)
		bp := s.Best()
		suite.Nil(err)
		suite.Equal(p1.IP, bp.IP)
	}
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
