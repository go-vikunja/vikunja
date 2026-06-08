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

package auth

import (
	"context"
	"testing"

	"code.vikunja.io/veans/internal/client"
)

func TestAcquireHumanToken_TokenShortCircuit(t *testing.T) {
	// When opts.Token is set, no prompts and no HTTP calls happen — the
	// nil client confirms that nothing tries to dial out.
	tok, err := AcquireHumanToken(context.Background(), (*client.Client)(nil), LoginOptions{Token: "abc"}, &recordingPrompter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "abc" {
		t.Fatalf("got %q, want abc", tok)
	}
}

type recordingPrompter struct {
	calls []string
}

func (r *recordingPrompter) ReadLine(p string) (string, error) {
	r.calls = append(r.calls, "line:"+p)
	return "", nil
}

func (r *recordingPrompter) ReadPassword(p string) (string, error) {
	r.calls = append(r.calls, "pw:"+p)
	return "", nil
}
