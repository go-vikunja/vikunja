// Package credentials handles bot-token storage with a keychain → env → file
// fallback chain. The store is keyed by (server, account); `account` is the
// bot's username — the human's token is never persisted.
package credentials

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when no backend has the requested credential.
var ErrNotFound = errors.New("credential not found")

// Store is the read/write contract every backend implements.
type Store interface {
	Get(server, account string) (string, error)
	Set(server, account, token string) error
	Delete(server, account string) error
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
func (c *Chain) Set(server, account, token string) error {
	for _, b := range c.Backends {
		if _, ok := b.(*EnvBackend); ok {
			continue
		}
		if err := b.Set(server, account, token); err == nil {
			return nil
		} else if !errors.Is(err, errReadOnly) {
			return fmt.Errorf("%s: %w", b.Name(), err)
		}
	}
	return errors.New("no writable backend available")
}

// Delete removes from every writable backend (best-effort).
func (c *Chain) Delete(server, account string) error {
	var firstErr error
	for _, b := range c.Backends {
		if _, ok := b.(*EnvBackend); ok {
			continue
		}
		if err := b.Delete(server, account); err != nil && !errors.Is(err, ErrNotFound) && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
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
