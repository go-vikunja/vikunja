package credentials

import (
	"errors"
	"path/filepath"
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

	if err := b.Delete("https://example.com", "bot-foo"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := b.Get("https://example.com", "bot-foo"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := b.Get("https://example.com", "bot-bar"); err != nil {
		t.Fatalf("bar should still exist: %v", err)
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
func (s *stubBackend) Delete(server, account string) error {
	k := server + "::" + account
	if _, ok := s.store[k]; !ok {
		return ErrNotFound
	}
	delete(s.store, k)
	return nil
}
