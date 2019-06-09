// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package storage

type SelectOptions struct {
	Filters []Filter
	Limit   int
	Offset  int
}

type SelectOption func(*SelectOptions)

// WithFilter adds a filter function to the list of filters used during the Select call
func WithFilter(fn ...Filter) SelectOption {
	return func(options *SelectOptions) {
		options.Filters = append(options.Filters, fn...)
	}
}

// WithLimit sets the limit number of proxies returned
func WithLimit(limit int) SelectOption {
	return func(options *SelectOptions) {
		options.Limit = limit
	}
}

// WithLimit sets the proxies return from offset position in storage
func WithOffset(offset int) SelectOption {
	return func(options *SelectOptions) {
		options.Offset = offset
	}
}
