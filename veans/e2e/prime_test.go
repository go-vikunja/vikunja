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
