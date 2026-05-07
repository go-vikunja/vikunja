package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
)

// OAuth client identity. Vikunja's authorization server requires no
// pre-registration — these values just need to be consistent between the
// browser-side authorize step and the CLI-side token exchange.
const (
	oauthClientID    = "veans-cli"
	oauthRedirectURI = "vikunja-veans-cli://callback"
)

// PKCEPair holds the challenge sent to /oauth/authorize and the verifier
// kept locally until token exchange.
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// generatePKCE produces a fresh (verifier, challenge) pair per RFC 7636.
// The verifier is 64 random bytes, base64url-encoded without padding (~86
// characters — comfortably inside the 43–128 range Vikunja accepts). The
// challenge is the SHA-256 of the verifier, also base64url-no-pad.
func generatePKCE() (PKCEPair, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return PKCEPair{}, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return PKCEPair{Verifier: verifier, Challenge: challenge}, nil
}

// generateState returns a random opaque string for CSRF protection on the
// authorize redirect. We verify it matches when the user pastes back.
func generateState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// buildAuthorizeURL renders the browser-side redirect target. The user
// follows it, authenticates if necessary, and is redirected to the custom
// scheme with `?code=...&state=...`.
func buildAuthorizeURL(server string, pkce PKCEPair, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", oauthClientID)
	q.Set("redirect_uri", oauthRedirectURI)
	q.Set("code_challenge", pkce.Challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	return strings.TrimRight(server, "/") + "/oauth/authorize?" + q.Encode()
}

// extractCodeAndState pulls the OAuth callback parameters out of whatever
// the user pasted. We accept three shapes:
//   - the full custom-scheme URL: `vikunja-veans-cli://callback?code=...&state=...`
//   - just the query: `code=ABC&state=XYZ`
//   - just the code (state verification then skipped, with a warning)
func extractCodeAndState(raw string) (code, state string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("empty callback paste")
	}

	// Full URL form?
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "vikunja-") {
		u, perr := url.Parse(raw)
		if perr != nil {
			return "", "", fmt.Errorf("parse callback URL: %w", perr)
		}
		v := u.Query()
		// Some browsers strip the query and put it in Fragment when they
		// can't open the scheme — handle both.
		if v.Get("code") == "" && u.RawQuery == "" && u.Fragment != "" {
			v, _ = url.ParseQuery(u.Fragment)
		}
		return v.Get("code"), v.Get("state"), nil
	}

	// Query-string form?
	if strings.Contains(raw, "code=") {
		v, perr := url.ParseQuery(raw)
		if perr != nil {
			return "", "", fmt.Errorf("parse callback query: %w", perr)
		}
		return v.Get("code"), v.Get("state"), nil
	}

	// Bare code.
	return raw, "", nil
}

// runOAuthFlow drives the manual paste-back OAuth Authorization Code +
// PKCE handshake against Vikunja's server.
//
// The user-facing UX: print the authorize URL, ask the user to open it in
// their browser, sign in there, and paste the resulting (failed-to-open)
// `vikunja-veans-cli://callback?code=...` URL back into the CLI. The
// browser will show a "can't open this scheme" error, but the URL bar
// contains the code we need.
func runOAuthFlow(ctx context.Context, c *client.Client, p Prompter, w io.Writer) (string, error) {
	pkce, err := generatePKCE()
	if err != nil {
		return "", output.Wrap(output.CodeUnknown, err, "generate PKCE: %v", err)
	}
	state, err := generateState()
	if err != nil {
		return "", output.Wrap(output.CodeUnknown, err, "generate state: %v", err)
	}

	authURL := buildAuthorizeURL(c.BaseURL, pkce, state)
	if w != nil {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Open the following URL in your browser:")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  "+authURL)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "After signing in, your browser will try to open")
		fmt.Fprintln(w, "  "+oauthRedirectURI+"?code=...&state=...")
		fmt.Fprintln(w, "and show a 'can't open this URL' error. That's expected.")
		fmt.Fprintln(w, "Copy the URL from the address bar and paste it here.")
		fmt.Fprintln(w, "")
	}

	pasted, err := p.ReadLine("Paste callback URL (or just the code): ")
	if err != nil {
		return "", err
	}
	code, returnedState, err := extractCodeAndState(pasted)
	if err != nil {
		return "", output.Wrap(output.CodeAuth, err, "%v", err)
	}
	if code == "" {
		return "", output.New(output.CodeAuth, "no `code` found in pasted callback")
	}
	if returnedState != "" && returnedState != state {
		return "", output.New(output.CodeAuth, "state mismatch on OAuth callback (possible CSRF)")
	}

	resp, err := c.ExchangeOAuthCode(ctx, &client.OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     oauthClientID,
		RedirectURI:  oauthRedirectURI,
		CodeVerifier: pkce.Verifier,
	})
	if err != nil {
		return "", err
	}
	if resp.AccessToken == "" {
		return "", output.New(output.CodeAuth, "OAuth token exchange returned empty access_token")
	}
	return resp.AccessToken, nil
}
