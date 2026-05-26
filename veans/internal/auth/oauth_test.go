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
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestBuildAuthorizeURL(t *testing.T) {
	u := buildAuthorizeURL(
		"https://vikunja.example.com",
		"http://127.0.0.1:54321/callback",
		PKCEPair{Challenge: "CHL"},
		"S",
	)
	if !strings.HasPrefix(u, "https://vikunja.example.com/oauth/authorize?") {
		t.Fatalf("unexpected prefix: %s", u)
	}
	for _, want := range []string{
		"response_type=code",
		"client_id=" + oauthClientID,
		"code_challenge=CHL",
		"code_challenge_method=S256",
		"state=S",
		// redirect_uri carried through (URL-encoded)
		"redirect_uri=http%3A%2F%2F127.0.0.1%3A54321%2Fcallback",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("authorize URL missing %q: %s", want, u)
		}
	}
	// Server URL with trailing slash should still produce a single slash
	// before the path.
	u2 := buildAuthorizeURL("https://vikunja.example.com/", "", PKCEPair{}, "")
	if strings.Contains(u2, "//oauth") {
		t.Errorf("trailing slash leaked into URL: %s", u2)
	}
}

func TestGenerateState_Shape(t *testing.T) {
	s, err := generateState()
	if err != nil {
		t.Fatal(err)
	}
	// 24 random bytes base64url-no-pad → 32 chars.
	if len(s) < 30 {
		t.Fatalf("state length %d shorter than expected", len(s))
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z',
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
		default:
			t.Fatalf("state contains illegal rune %q", r)
		}
	}
	// Decodes cleanly as base64url-no-pad.
	if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
		t.Fatalf("state isn't base64url-no-pad: %v", err)
	}
	// Two consecutive states should differ — sanity for entropy.
	s2, _ := generateState()
	if s == s2 {
		t.Fatal("two consecutive states are identical — entropy is broken")
	}
}

// newCallbackHandler returns just the http.Handler portion of
// newCallbackServer so tests can drive it directly with httptest.NewRecorder
// without binding a real loopback socket.
func newCallbackHandler(t *testing.T) (http.Handler, <-chan callbackResult) {
	t.Helper()
	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server, ch := newCallbackServer(listener)
	return server.Handler, ch
}

func TestNewCallbackServer_HappyPath(t *testing.T) {
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/callback?code=abc&state=xyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	select {
	case res := <-ch:
		if res.code != "abc" {
			t.Errorf("code = %q, want abc", res.code)
		}
		if res.state != "xyz" {
			t.Errorf("state = %q, want xyz", res.state)
		}
		if res.err != nil {
			t.Errorf("err = %v, want nil", res.err)
		}
	default:
		t.Fatal("no result pushed to channel")
	}

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html…", ct)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNewCallbackServer_AuthzServerError(t *testing.T) {
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodGet,
		"/callback?error=access_denied&error_description=user+declined", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	select {
	case res := <-ch:
		if res.err == nil {
			t.Fatal("err = nil, want non-nil")
		}
		// renderCallbackPage uses error_description when present; the
		// handler also stuffs it into res.err. "user declined" comes
		// straight from error_description.
		if !strings.Contains(res.err.Error(), "user declined") {
			t.Errorf("err = %q, want it to mention error_description", res.err.Error())
		}
	default:
		t.Fatal("no result pushed to channel")
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestNewCallbackServer_AuthzServerErrorOnlyCode(t *testing.T) {
	// When error_description is empty, the handler falls back to the
	// `error` code in the user-visible message.
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/callback?error=access_denied", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	select {
	case res := <-ch:
		if res.err == nil {
			t.Fatal("err = nil, want non-nil")
		}
		if !strings.Contains(res.err.Error(), "access_denied") {
			t.Errorf("err = %q, want it to mention error code", res.err.Error())
		}
	default:
		t.Fatal("no result pushed to channel")
	}
}

func TestNewCallbackServer_EmptyCode(t *testing.T) {
	// Empty `code` without an `error` parameter is the handler's job
	// only to forward verbatim — waitForCallback is the one that
	// upgrades that to an error.
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/callback?state=xyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	select {
	case res := <-ch:
		if res.code != "" {
			t.Errorf("code = %q, want empty", res.code)
		}
		if res.state != "xyz" {
			t.Errorf("state = %q, want xyz", res.state)
		}
		if res.err != nil {
			t.Errorf("err = %v, want nil", res.err)
		}
	default:
		t.Fatal("no result pushed to channel")
	}
}

func TestNewCallbackServer_MethodNotAllowed(t *testing.T) {
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/callback?code=abc&state=xyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
	if a := rec.Header().Get("Allow"); a != "GET" {
		t.Errorf("Allow header = %q, want GET", a)
	}
	select {
	case res := <-ch:
		t.Fatalf("nothing should be pushed for a rejected method, got %+v", res)
	default:
	}
}

func TestNewCallbackServer_WrongPath(t *testing.T) {
	handler, ch := newCallbackHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/something-else", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	select {
	case res := <-ch:
		t.Fatalf("nothing should be pushed for a 404, got %+v", res)
	default:
	}
}

func TestRenderCallbackPage_HTMLEscapesError(t *testing.T) {
	rec := httptest.NewRecorder()
	renderCallbackPage(rec, &fakeError{msg: `<script>alert(1)</script>`})

	body := rec.Body.String()
	if strings.Contains(body, "<script>") {
		t.Errorf("body contains raw <script>: %s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Errorf("body missing escaped script tag: %s", body)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 on error", rec.Code)
	}
}

func TestRenderCallbackPage_SuccessNoErrorPath(t *testing.T) {
	rec := httptest.NewRecorder()
	renderCallbackPage(rec, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "veans is authorized") {
		t.Errorf("missing success message: %s", rec.Body.String())
	}
}

type fakeError struct{ msg string }

func (f *fakeError) Error() string { return f.msg }
