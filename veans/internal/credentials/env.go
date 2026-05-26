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

func (*EnvBackend) Set(_, _, _ string) error    { return errReadOnly }
func (*EnvBackend) Delete(_, _ string) error    { return errReadOnly }
