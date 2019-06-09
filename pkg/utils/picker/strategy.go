// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package picker

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// Next is a function that returns the next element based on the strategy
type Next func() (interface{}, error)

// Strategy is a selection strategy e.g random, round robin
type Strategy func([]interface{}) Next

var ErrNoneAvailable = errors.New("none available")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random is a random strategy algorithm for element selection
func Random(elements []interface{}) Next {
	return func() (interface{}, error) {
		if len(elements) == 0 {
			return nil, ErrNoneAvailable
		}

		i := rand.Int() % len(elements)
		return elements[i], nil
	}
}

// RoundRobin is a round robin strategy algorithm for element selection
func RoundRobin(elements []interface{}) Next {
	var i = rand.Int()
	var mu sync.Mutex

	return func() (interface{}, error) {
		if len(elements) == 0 {
			return nil, ErrNoneAvailable
		}

		mu.Lock()
		elem := elements[i%len(elements)]
		i++
		mu.Unlock()

		return elem, nil
	}
}
