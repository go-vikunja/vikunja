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

package bootstrap

import (
	"errors"
	"path/filepath"
	"testing"

	"code.vikunja.io/veans/internal/trust"
)

func newTestGate(t *testing.T, p trust.Prompter) *trust.Gate {
	t.Helper()
	t.Setenv("VEANS_SERVER", "")
	return &trust.Gate{Store: trust.NewStore(filepath.Join(t.TempDir(), "trusted.json")), Prompter: p}
}

func TestTrustServer_RecordsTypedServerWithoutAsking(t *testing.T) {
	for _, tc := range []struct{ typed, canonical string }{
		{"https://vikunja.example.com", "https://vikunja.example.com"},
		{"vikunja.example.com", "https://vikunja.example.com:3456"},
		{"localhost", "http://localhost:3456"},
	} {
		p := &queuePrompter{answers: []string{"n"}}
		gate := newTestGate(t, p)
		if err := trustServer(gate, tc.typed, tc.canonical); err != nil {
			t.Fatalf("%q -> %q: %v", tc.typed, tc.canonical, err)
		}
		if p.calls != 0 {
			t.Fatalf("%q -> %q: prompted", tc.typed, tc.canonical)
		}
		gate.Prompter = nil
		if err := gate.Check(tc.canonical); err != nil {
			t.Fatalf("%q -> %q: not recorded: %v", tc.typed, tc.canonical, err)
		}
	}
}

func TestTrustServer_SchemeChangeAsks(t *testing.T) {
	p := &queuePrompter{answers: []string{"n", "y"}}
	gate := newTestGate(t, p)
	err := trustServer(gate, "http://vikunja.example.com", "https://vikunja.example.com")
	if !errors.Is(err, trust.ErrUntrusted) {
		t.Fatalf("got %v, want ErrUntrusted", err)
	}
	if p.calls != 1 {
		t.Fatalf("asked %d times, want 1", p.calls)
	}

	if err := trustServer(gate, "http://vikunja.example.com", "https://vikunja.example.com"); err != nil {
		t.Fatal(err)
	}
}
