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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
)

// oauthClientID is what veans presents to Vikunja's authorization server.
// Vikunja's OAuth provider doesn't require client registration — the value
// just needs to be consistent across the authorize and token-exchange steps.
const oauthClientID = "veans-cli"

// loopbackTimeout caps how long we wait for the user to complete the
// browser-side handshake before giving up.
const loopbackTimeout = 5 * time.Minute

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

// generateState returns a random opaque string for CSRF protection.
func generateState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// buildAuthorizeURL renders the browser-side redirect target.
func buildAuthorizeURL(server, redirectURI string, pkce PKCEPair, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", oauthClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", pkce.Challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	return strings.TrimRight(server, "/") + "/oauth/authorize?" + q.Encode()
}

// callbackResult carries the parsed query parameters from the loopback
// callback request, or any error that prevented a clean handshake.
type callbackResult struct {
	code  string
	state string
	err   error
}

// runOAuthFlow drives an OAuth Authorization Code + PKCE handshake against
// Vikunja's server using a localhost loopback listener (RFC 8252):
// bind 127.0.0.1:0, open the authorize URL in the browser, capture the
// callback, exchange the code for a token.
//
// The prompter is retained on the signature for symmetry with the
// password flow but isn't called — the loopback handshake completes
// without further user input beyond the in-browser sign-in.
func runOAuthFlow(ctx context.Context, c *client.Client, _ Prompter, w io.Writer) (string, error) {
	pkce, err := generatePKCE()
	if err != nil {
		return "", output.Wrap(output.CodeUnknown, err, "generate PKCE: %v", err)
	}
	state, err := generateState()
	if err != nil {
		return "", output.Wrap(output.CodeUnknown, err, "generate state: %v", err)
	}

	listener, redirectURI, err := bindLoopbackListener(ctx)
	if err != nil {
		return "", err
	}

	server, resultCh := newCallbackServer(listener)
	go func() { _ = server.Serve(listener) }()
	// Shutdown uses a detached context derived from ctx so cancellation
	// at the outer scope still allows the graceful-stop to drain.
	shutdownParent := context.WithoutCancel(ctx)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(shutdownParent, 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	authURL := buildAuthorizeURL(c.BaseURL, redirectURI, pkce, state)
	announceBrowserStep(w, authURL)
	openBrowser(ctx, authURL)

	result, err := waitForCallback(ctx, resultCh)
	if err != nil {
		return "", err
	}
	if result.state != state {
		return "", output.New(output.CodeAuth,
			"state mismatch on OAuth callback (possible CSRF)")
	}

	resp, err := c.ExchangeOAuthCode(ctx, &client.OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         result.code,
		ClientID:     oauthClientID,
		RedirectURI:  redirectURI,
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

// bindLoopbackListener picks a free port on 127.0.0.1 and returns a
// listener + the corresponding `http://127.0.0.1:NNN/callback` URI for
// the OAuth `redirect_uri` parameter.
func bindLoopbackListener(ctx context.Context) (net.Listener, string, error) {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", output.Wrap(output.CodeUnknown, err,
			"bind loopback port for OAuth callback: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	return listener, fmt.Sprintf("http://127.0.0.1:%d/callback", port), nil
}

// newCallbackServer returns an http.Server bound to `listener` whose
// /callback handler parses the authorization-server redirect query and
// pushes the result onto the returned channel.
func newCallbackServer(listener net.Listener) (*http.Server, <-chan callbackResult) {
	resultCh := make(chan callbackResult, 1)
	server := &http.Server{
		Addr:              listener.Addr().String(),
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(rw, r)
				return
			}
			q := r.URL.Query()
			res := callbackResult{code: q.Get("code"), state: q.Get("state")}
			if errCode := q.Get("error"); errCode != "" {
				desc := q.Get("error_description")
				if desc == "" {
					desc = errCode
				}
				res.err = fmt.Errorf("authorization failed: %s", desc)
			}
			renderCallbackPage(rw, res.err)
			select {
			case resultCh <- res:
			default:
			}
		}),
	}
	return server, resultCh
}

// waitForCallback blocks until the loopback handler fires, ctx cancels,
// or loopbackTimeout elapses.
func waitForCallback(ctx context.Context, resultCh <-chan callbackResult) (callbackResult, error) {
	timer := time.NewTimer(loopbackTimeout)
	defer timer.Stop()
	select {
	case result := <-resultCh:
		if result.err != nil {
			return result, output.Wrap(output.CodeAuth, result.err, "%v", result.err)
		}
		if result.code == "" {
			return result, output.New(output.CodeAuth, "no `code` returned from OAuth callback")
		}
		return result, nil
	case <-timer.C:
		return callbackResult{}, output.New(output.CodeAuth,
			"OAuth flow timed out after %s — re-run init with --use-password or --token", loopbackTimeout)
	case <-ctx.Done():
		return callbackResult{}, ctx.Err()
	}
}

func announceBrowserStep(w io.Writer, authURL string) {
	if w == nil {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opening your browser to authorize veans:")
	fmt.Fprintln(w, "  "+authURL)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If the browser doesn't open, paste the URL above manually.")
	fmt.Fprintln(w)
}

// renderCallbackPage writes a minimal HTML response to the user's browser
// after the loopback callback fires. We don't ship any framework — a few
// lines of inlined HTML are enough to confirm completion.
func renderCallbackPage(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `<!doctype html><html><body style="font-family:system-ui,sans-serif;max-width:32rem;margin:4rem auto;padding:0 1rem">
<h1>veans: authorization failed</h1>
<p>%s</p>
<p>You can close this tab and re-run <code>veans init</code>.</p>
</body></html>`, err.Error())
		return
	}
	_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui,sans-serif;max-width:32rem;margin:4rem auto;padding:0 1rem">
<h1>veans is authorized</h1>
<p>You can close this tab and return to the terminal.</p>
</body></html>`))
}

// openBrowser tries to launch the user's default browser at `url`. Failure
// is ignored — the calling flow already prints the URL to stderr so the
// user can open it themselves.
func openBrowser(ctx context.Context, url string) {
	_ = osOpen(ctx, url)
}

// silence the unused-import linter when errors isn't referenced elsewhere.
var _ = errors.New
