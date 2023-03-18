// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"testing"
)

func TestParseAnonymity(t *testing.T) {
	tests := []struct {
		name string
		ss   []string
		want Anonymity
	}{
		{
			name: "Unknown",
			ss:   []string{"asdasd", "", "Elitx"},
			want: AnonymityUnknown,
		},
		{
			name: "Elite",
			ss:   []string{"elite ", "高匿", "高匿名"},
			want: Elite,
		},
		{
			name: "Anonymous",
			ss:   []string{" anonymous", "普通匿名", "普匿"},
			want: Anonymous,
		},
		{
			name: "Transparent",
			ss:   []string{"透明", "TRANSPARENT"},
			want: Transparent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, s := range tt.ss {
				got := ParseAnonymity(s)
				if got != tt.want {
					t.Errorf("ParseAnonymity s = %s, got = %+v, want = %+v", s, got, tt.want)
				}
			}
		})
	}
}
