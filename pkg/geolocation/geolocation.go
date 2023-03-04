// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package geolocation

import (
	"context"
)

// Geolocation includes Geographical related information, country, ISP, latitude and longitude, etc.
type Geolocation struct {
	CountryName string
	CountryCode string
	RegionName  string
	RegionCode  string
	City        string
	Lat         float32
	Lon         float32
	ISP         string
	Org         string
}

// GeolocatorName describes the name of the Geolocator.
type GeolocatorName string

// Geolocator locate the ip geo location.
type Geolocator interface {
	Name() GeolocatorName
	Locate(ctx context.Context, ip string) (*Geolocation, error)
}
