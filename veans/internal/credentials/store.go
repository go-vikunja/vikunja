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

// Package credentials handles bot-token storage with a keychain → env → file
// fallback chain. The store is keyed by (server, account); `account` is the
// bot's username — the human's token is never persisted.
package credentials

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ChainStderr is the writer the Chain uses for operator-visible warnings
// (currently: backend-fallthrough notices on Set). Tests override it; in
// production it points at os.Stderr.
var ChainStderr io.Writer = os.Stderr

// ErrNotFound is returned when no backend has the requested credential.
var ErrNotFound = errors.New("credential not found")

// Store is the read/write contract every backend implements.
type Store interface {
	Get(server, account string) (string, error)
	Set(server, account, token string) error
	// Name is used in error messages.
	Name() string
}

// Chain queries each backend in order on Get; writes go to the first writable
// backend. Env (read-only) is skipped on writes. The order is keychain →
// env → file, matching the plan.
type Chain struct {
	Backends []Store
}

func (c *Chain) Name() string { return "chain" }

// Get returns the first non-NotFound result from any backend.
func (c *Chain) Get(server, account string) (string, error) {
	var lastErr error
	for _, b := range c.Backends {
		tok, err := b.Get(server, account)
		if err == nil {
			return tok, nil
		}
		if !errors.Is(err, ErrNotFound) {
			lastErr = fmt.Errorf("%s: %w", b.Name(), err)
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", ErrNotFound
}

// Set writes to the first backend that accepts a write. Env is read-only.
// Backends that error out (e.g. keyring on a host with no dbus) are skipped
// transparently, falling through to the next — the file backend is the
// reliable last-resort. Only if every writable backend fails do we surface
// the last error.
//
// When a write fails on one writable backend and a later one succeeds, a
// single-line warning is printed to ChainStderr naming both backends.
// This is observability for the silent-shadow case: a stale keyring entry
// from a prior successful write can mask the freshly-written file token if
// keyring transiently rejects the new Set. The warning gives the operator
// a breadcrumb; Set itself still returns nil because the write landed
// somewhere durable.
func (c *Chain) Set(server, account, token string) error {
	var (
		lastErr    error
		failedName string
		failedErr  error
	)
	for _, b := range c.Backends {
		if _, ok := b.(*EnvBackend); ok {
			continue
		}
		err := b.Set(server, account, token)
		if err == nil {
			if failedName != "" {
				fmt.Fprintf(ChainStderr,
					"veans: credential store: %s rejected write (%v); falling back to %s\n",
					failedName, failedErr, b.Name())
			}
			return nil
		}
		if errors.Is(err, errReadOnly) {
			continue
		}
		// Remember the most recent non-readonly failure so a later success
		// can surface it, or so we can return it if every backend fails.
		failedName = b.Name()
		failedErr = err
		lastErr = fmt.Errorf("%s: %w", b.Name(), err)
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("no writable backend available")
}

// errReadOnly is sentinel for backends that refuse writes (env).
var errReadOnly = errors.New("read-only backend")

// Default builds the standard keychain → env → file chain.
func Default() *Chain {
	return &Chain{
		Backends: []Store{
			NewKeyringBackend(),
			NewEnvBackend(),
			NewFileBackend(""),
		},
	}
}
