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

package e2e

import (
	"strings"
	"testing"
)

// TestPrime_RendersWithProjectAndBot pins the literal anchors hooks depend
// on. Mirrors plan e2e test 12.
func TestPrime_RendersWithProjectAndBot(t *testing.T) {
	ws, h := provisionWorkspace(t)
	cfg := loadConfig(t, ws)

	out, _, code := h.Run(t, ws, "prime")
	if code != 0 {
		t.Fatalf("prime exit %d", code)
	}

	mustContain := []string{
		"<EXTREMELY_IMPORTANT>",
		cfg.Bot.Username,
		"Refs:",
		"veans claim",
		"Todo",
		"In Progress",
		"In Review",
		"Done",
		"Scrapped",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("prime output missing %q", s)
		}
	}
}

// TestPrime_SilentOutsideWorkspace verifies the safe-globally-installed
// hook contract: no .veans.yml ⇒ silent + exit 0.
func TestPrime_SilentOutsideWorkspace(t *testing.T) {
	h := New(t)

	// A workspace with no .veans.yml — just a temp dir.
	ws := h.NewWorkspace(t)

	stdout, stderr, code := h.Run(t, ws, "prime")
	if code != 0 {
		t.Fatalf("prime exit %d (expected 0)\n%s\n%s", code, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected silent stdout outside workspace, got: %s", stdout)
	}
}
