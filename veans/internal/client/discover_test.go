// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package client

import (
	"slices"
	"testing"
)

func TestServerCandidates(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "bare hostname → https first, then http, with default-port fallbacks",
			input: "vikunja.example.com",
			want: []string{
				"https://vikunja.example.com",
				"https://vikunja.example.com:3456",
				"http://vikunja.example.com",
				"http://vikunja.example.com:3456",
			},
		},
		{
			name:  "localhost defaults to http",
			input: "localhost",
			want: []string{
				"http://localhost",
				"http://localhost:3456",
				"https://localhost",
				"https://localhost:3456",
			},
		},
		{
			name:  "user-supplied /api/v1 suffix is trimmed (so the probe doesn't double it up)",
			input: "https://vikunja.example.com/api/v1",
			want: []string{
				"https://vikunja.example.com",
				"https://vikunja.example.com:3456",
				"http://vikunja.example.com",
				"http://vikunja.example.com:3456",
			},
		},
		{
			name:  "explicit port is respected — no default-port fallback added",
			input: "https://vikunja.example.com:8443",
			want: []string{
				"https://vikunja.example.com:8443",
				"http://vikunja.example.com:8443",
			},
		},
		{
			name:  "subpath install keeps the prefix",
			input: "https://example.com/vikunja",
			want: []string{
				"https://example.com/vikunja",
				"https://example.com:3456/vikunja",
				"http://example.com/vikunja",
				"http://example.com:3456/vikunja",
			},
		},
		{
			name:  "127.0.0.1 with default port (common dev setup)",
			input: "127.0.0.1:3456",
			want: []string{
				"http://127.0.0.1:3456",
				"https://127.0.0.1:3456",
			},
		},
		{
			name:  "trailing slash trimmed",
			input: "https://vikunja.example.com/",
			want: []string{
				"https://vikunja.example.com",
				"https://vikunja.example.com:3456",
				"http://vikunja.example.com",
				"http://vikunja.example.com:3456",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := serverCandidates(c.input)
			if err != nil {
				t.Fatalf("serverCandidates(%q): %v", c.input, err)
			}
			if !slices.Equal(got, c.want) {
				t.Errorf("serverCandidates(%q):\n  got  %v\n  want %v", c.input, got, c.want)
			}
		})
	}
}

func TestServerCandidates_EmptyInput(t *testing.T) {
	// "" is the only input shape DiscoverServer rejects at the entry
	// (before reaching serverCandidates). The lower-level helper itself
	// reports "missing host" through the url.Parse path.
	if _, err := serverCandidates(""); err == nil {
		t.Error("empty input should error")
	}
}
