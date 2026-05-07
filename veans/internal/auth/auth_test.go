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
