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

package credentials

import "os"

// EnvBackend is read-only. VEANS_TOKEN is intended for CI / containers where
// the keychain is unavailable and writing a credentials file is undesirable.
//
// VEANS_TOKEN matches any (server, account) lookup — there's only one slot.
// VEANS_SERVER, when set, additionally pins the server it applies to.
type EnvBackend struct{}

func NewEnvBackend() *EnvBackend { return &EnvBackend{} }
func (*EnvBackend) Name() string { return "env" }

func (*EnvBackend) Get(server, _ string) (string, error) {
	tok := os.Getenv("VEANS_TOKEN")
	if tok == "" {
		return "", ErrNotFound
	}
	if pinned := os.Getenv("VEANS_SERVER"); pinned != "" && pinned != server {
		return "", ErrNotFound
	}
	return tok, nil
}

func (*EnvBackend) Set(_, _, _ string) error { return errReadOnly }
func (*EnvBackend) Delete(_, _ string) error { return errReadOnly }
