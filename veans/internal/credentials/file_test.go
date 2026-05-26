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

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestFileBackend_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.yml")
	b := NewFileBackend(path)

	if _, err := b.Get("https://example.com", "bot-foo"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := b.Set("https://example.com", "bot-foo", "tok-123"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	tok, err := b.Get("https://example.com", "bot-foo")
	if err != nil {
		t.Fatalf("Get after Set: %v", err)
	}
	if tok != "tok-123" {
		t.Fatalf("got %q, want tok-123", tok)
	}

	// Update in place.
	if err := b.Set("https://example.com", "bot-foo", "tok-456"); err != nil {
		t.Fatalf("Set update: %v", err)
	}
	tok, _ = b.Get("https://example.com", "bot-foo")
	if tok != "tok-456" {
		t.Fatalf("update lost: got %q", tok)
	}

	// Different account — separate row.
	if err := b.Set("https://example.com", "bot-bar", "tok-789"); err != nil {
		t.Fatalf("Set bar: %v", err)
	}
	tokBar, _ := b.Get("https://example.com", "bot-bar")
	if tokBar != "tok-789" {
		t.Fatalf("bar got %q", tokBar)
	}
}

func TestChain_FallsThroughOnNotFound(t *testing.T) {
	dir := t.TempDir()
	file := NewFileBackend(filepath.Join(dir, "credentials.yml"))
	stub := &stubBackend{store: map[string]string{}}
	c := &Chain{Backends: []Store{stub, file}}

	// First backend has nothing; second is empty too.
	if _, err := c.Get("s", "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Set should write to the first writable backend (stub here).
	if err := c.Set("s", "a", "tok"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if stub.store["s::a"] != "tok" {
		t.Fatalf("expected stub to receive write")
	}

	// Get should now find it via stub.
	if got, _ := c.Get("s", "a"); got != "tok" {
		t.Fatalf("got %q want tok", got)
	}
}

type stubBackend struct {
	store map[string]string
}

func (s *stubBackend) Name() string { return "stub" }
func (s *stubBackend) Get(server, account string) (string, error) {
	if v, ok := s.store[server+"::"+account]; ok {
		return v, nil
	}
	return "", ErrNotFound
}
func (s *stubBackend) Set(server, account, token string) error {
	s.store[server+"::"+account] = token
	return nil
}

// failingBackend always errors on Set with a non-readonly, non-NotFound error,
// simulating e.g. a keyring with no dbus available. Get always reports
// ErrNotFound so the chain's Get path stays uninteresting for these tests.
type failingBackend struct {
	name string
	err  error
}

func (f *failingBackend) Name() string                    { return f.name }
func (f *failingBackend) Get(_, _ string) (string, error) { return "", ErrNotFound }
func (f *failingBackend) Set(_, _, _ string) error        { return f.err }

// TestFileBackend_SetReassertsMode covers the os.Chmod(path, 0o600) at the
// end of save: a pre-existing file at a wider mode must be narrowed.
func TestFileBackend_SetReassertsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.yml")
	// Pre-create the file at 0o644 so Rename onto the existing inode can
	// (on some filesystems) preserve the wider mode.
	if err := os.WriteFile(path, []byte("credentials: []\n"), 0o644); err != nil { //nolint:gosec // test fixture: intentionally wider than 0600
		t.Fatalf("seed file: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod seed: %v", err)
	}

	b := NewFileBackend(path)
	if err := b.Set("https://example.com", "bot-foo", "tok-123"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode after Set: got %o, want 0600", perm)
	}
}

// TestFileBackend_SetCleansUpTmpFile asserts that the atomic-write tmp file
// (.credentials-*.tmp) is renamed away — no stray tmp should remain after a
// successful Set.
func TestFileBackend_SetCleansUpTmpFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.yml")
	b := NewFileBackend(path)

	if err := b.Set("https://example.com", "bot-foo", "tok-123"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".credentials-") && strings.HasSuffix(name, ".tmp") {
			t.Fatalf("leftover tmp file after Set: %s", name)
		}
	}
}

// TestFileBackend_ConcurrentWritersSerialize fans two goroutines into Set
// with different (server, account) keys. The flock should serialize the
// load → mutate → save sequence so both entries are persisted, even
// though either could otherwise stomp on the other's load snapshot.
func TestFileBackend_ConcurrentWritersSerialize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.yml")
	b := NewFileBackend(path)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := b.Set("https://a.example.com", "bot-a", "tok-a"); err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := b.Set("https://b.example.com", "bot-b", "tok-b"); err != nil {
			errCh <- err
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent Set: %v", err)
	}

	gotA, err := b.Get("https://a.example.com", "bot-a")
	if err != nil || gotA != "tok-a" {
		t.Fatalf("a: got %q err=%v, want tok-a", gotA, err)
	}
	gotB, err := b.Get("https://b.example.com", "bot-b")
	if err != nil || gotB != "tok-b" {
		t.Fatalf("b: got %q err=%v, want tok-b", gotB, err)
	}
}

// TestChain_SetWarnsOnFallback asserts that when an earlier writable backend
// errors and a later one succeeds, the chain writes a one-line warning to
// ChainStderr naming both backends. Set itself still returns nil because
// the write landed durably on the later backend.
func TestChain_SetWarnsOnFallback(t *testing.T) {
	dir := t.TempDir()
	file := NewFileBackend(filepath.Join(dir, "credentials.yml"))
	failing := &failingBackend{name: "keyring-stub", err: errors.New("dbus unavailable")}

	var buf bytes.Buffer
	origStderr := ChainStderr
	ChainStderr = &buf
	t.Cleanup(func() { ChainStderr = origStderr })

	c := &Chain{Backends: []Store{failing, file}}
	if err := c.Set("https://example.com", "bot-foo", "tok-xyz"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// The token must have landed in the file backend.
	if got, err := file.Get("https://example.com", "bot-foo"); err != nil || got != "tok-xyz" {
		t.Fatalf("file Get: got %q err=%v, want tok-xyz", got, err)
	}

	out := buf.String()
	if out == "" {
		t.Fatalf("expected warning on ChainStderr, got nothing")
	}
	if !strings.Contains(out, failing.Name()) {
		t.Fatalf("warning missing failing backend name %q: %s", failing.Name(), out)
	}
	if !strings.Contains(out, file.Name()) {
		t.Fatalf("warning missing fallback backend name %q: %s", file.Name(), out)
	}
}
