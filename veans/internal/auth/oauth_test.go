package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestGeneratePKCE_VerifierShape(t *testing.T) {
	pair, err := generatePKCE()
	if err != nil {
		t.Fatal(err)
	}
	// RFC 7636 §4.1: verifier is 43–128 chars, [A-Za-z0-9-._~].
	if len(pair.Verifier) < 43 || len(pair.Verifier) > 128 {
		t.Fatalf("verifier length %d out of [43,128]", len(pair.Verifier))
	}
	for _, r := range pair.Verifier {
		switch {
		case r >= 'A' && r <= 'Z',
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '-', r == '.', r == '_', r == '~':
		default:
			t.Fatalf("verifier contains illegal rune %q", r)
		}
	}
	// Challenge must be SHA256(verifier) base64url-no-pad.
	want := sha256.Sum256([]byte(pair.Verifier))
	got, err := base64.RawURLEncoding.DecodeString(pair.Challenge)
	if err != nil {
		t.Fatalf("challenge isn't base64url-no-pad: %v", err)
	}
	if string(got) != string(want[:]) {
		t.Fatal("challenge != SHA256(verifier)")
	}
}

func TestGeneratePKCE_Unique(t *testing.T) {
	a, _ := generatePKCE()
	b, _ := generatePKCE()
	if a.Verifier == b.Verifier {
		t.Fatal("two consecutive verifiers are identical — entropy is broken")
	}
}

func TestExtractCodeAndState_FullURL(t *testing.T) {
	code, state, err := extractCodeAndState("vikunja-veans-cli://callback?code=ABC123&state=XYZ")
	if err != nil {
		t.Fatal(err)
	}
	if code != "ABC123" || state != "XYZ" {
		t.Fatalf("got code=%q state=%q", code, state)
	}
}

func TestExtractCodeAndState_QueryOnly(t *testing.T) {
	code, state, err := extractCodeAndState("code=ABC&state=XYZ")
	if err != nil {
		t.Fatal(err)
	}
	if code != "ABC" || state != "XYZ" {
		t.Fatalf("got code=%q state=%q", code, state)
	}
}

func TestExtractCodeAndState_BareCode(t *testing.T) {
	code, state, err := extractCodeAndState("plain-code-value")
	if err != nil {
		t.Fatal(err)
	}
	if code != "plain-code-value" || state != "" {
		t.Fatalf("got code=%q state=%q", code, state)
	}
}

func TestExtractCodeAndState_EmptyError(t *testing.T) {
	if _, _, err := extractCodeAndState("   "); err == nil {
		t.Fatal("expected error on empty paste")
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	u := buildAuthorizeURL("https://vikunja.example.com", PKCEPair{Challenge: "CHL"}, "S")
	if !strings.HasPrefix(u, "https://vikunja.example.com/oauth/authorize?") {
		t.Fatalf("unexpected prefix: %s", u)
	}
	for _, want := range []string{
		"response_type=code",
		"client_id=" + oauthClientID,
		"code_challenge=CHL",
		"code_challenge_method=S256",
		"state=S",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("authorize URL missing %q: %s", want, u)
		}
	}
	// Server URL with trailing slash should still produce a single slash
	// before the path.
	u2 := buildAuthorizeURL("https://vikunja.example.com/", PKCEPair{}, "")
	if strings.Contains(u2, "//oauth") {
		t.Errorf("trailing slash leaked into URL: %s", u2)
	}
}
