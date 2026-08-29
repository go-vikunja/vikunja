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

package planka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/utils"

	"github.com/c2h5oh/datasize"
)

// maxResponseBytes caps how much json a planka instance can make us buffer.
const maxResponseBytes = 64 << 20

const maxRedirects = 10

// One deadline for login + probes + logout: the check runs inside an http handler and is not retried.
const credentialCheckTimeout = 15 * time.Second

// client talks to a Planka v2 instance. Exactly one of apiKey / jwt is set after login.
type client struct {
	baseURL string
	hc      *http.Client
	apiKey  string
	jwt     string
	// ownsJWT: the session was created by us via password login and must be deleted on logout.
	ownsJWT       bool
	currentUserID string
	// downloadHC follows the redirect planka answers with when attachments live on object storage.
	downloadHC *http.Client
}

// newClient normalises the url: trailing slashes and a trailing "/api" are stripped.
func newClient(rawURL string) (*client, error) {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return nil, &ErrInvalidConfig{}
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, &ErrNoPlankaAtURL{Reason: "invalid url"}
	}
	// userinfo would end up in log lines and error messages
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimRight(u.Path, "/")
	u.Path = strings.TrimSuffix(u.Path, "/api")
	u.Path = strings.TrimRight(u.Path, "/")

	hc := utils.NewSSRFSafeHTTPClient()
	hc.CheckRedirect = redirectPolicy(u)

	downloadHC := utils.NewSSRFSafeHTTPClient()
	downloadHC.CheckRedirect = downloadRedirectPolicy(u)

	return &client{
		baseURL:    u.String(),
		hc:         hc,
		downloadHC: downloadHC,
	}, nil
}

// redirectPolicy keeps credentials on the origin: go replays X-Api-Key, the accessToken cookie and the
// login body on redirects, so refuse another host and refuse dropping out of https.
func redirectPolicy(origin *url.URL) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if !sameHost(req.URL, origin) {
			return fmt.Errorf("%w: planka redirected to another host", migration.ErrRedirectRefused)
		}
		return checkDowngradeAndHops(req, via, origin)
	}
}

// downloadRedirectPolicy follows attachments to wherever they are stored, without the planka credentials:
// go only drops Cookie and Authorization when the hostname changes, and never drops X-Api-Key.
func downloadRedirectPolicy(origin *url.URL) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if err := checkDowngradeAndHops(req, via, origin); err != nil {
			return err
		}
		if !sameHost(req.URL, origin) {
			req.Header.Del("X-Api-Key")
			req.Header.Del("Cookie")
			req.Header.Del("Authorization")
		}
		return nil
	}
}

func checkDowngradeAndHops(req *http.Request, via []*http.Request, origin *url.URL) error {
	if origin.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("%w: planka redirected to a non-https url", migration.ErrRedirectRefused)
	}
	if len(via) >= maxRedirects {
		return fmt.Errorf("%w: too many redirects", migration.ErrRedirectRefused)
	}
	return nil
}

func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(hostWithoutDefaultPort(a), hostWithoutDefaultPort(b))
}

func hostWithoutDefaultPort(u *url.URL) string {
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		return u.Hostname()
	}
	return u.Host
}

type plankaUserResponse struct {
	Item plankaUser `json:"item"`
}

// login authenticates either with a token (API key or JWT) or with username + password.
func (c *client) login(ctx context.Context, token, username, password string) error {
	if token != "" {
		return c.loginWithToken(ctx, token)
	}
	if username != "" && password != "" {
		err := c.loginWithPassword(ctx, username, password)
		if err != nil {
			// the login can fail after the session was created, e.g. when a later probe errors
			c.logout(ctx)
		}
		return err
	}
	return &ErrInvalidConfig{}
}

// loginWithToken probes the token as API key first, then as Bearer JWT. Planka ignores
// X-Api-Key when an Authorization header is present, so both must never be sent together.
func (c *client) loginWithToken(ctx context.Context, token string) error {
	c.apiKey, c.jwt = token, ""
	me, status, err := c.probeMe(ctx)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return c.acceptUser(me)
	}

	c.apiKey, c.jwt = "", token
	me, status, err = c.probeMe(ctx)
	if err != nil {
		return err
	}
	if status == http.StatusOK {
		return c.acceptUser(me)
	}

	c.apiKey, c.jwt = "", ""
	return &ErrInvalidCredentials{}
}

// An empty id means something other than Planka answered 200.
func (c *client) acceptUser(me *plankaUserResponse) error {
	if me.Item.ID == "" {
		return &ErrNoPlankaAtURL{Reason: "/api/users/me returned no user"}
	}
	c.currentUserID = me.Item.ID
	return nil
}

// probeMe calls GET /api/users/me. Anything but 200 or 401 means there is no Planka api at the url.
// It bypasses the retrying helper: this runs while the api request of the user is waiting.
func (c *client) probeMe(ctx context.Context) (me *plankaUserResponse, status int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/users/me", nil)
	if err != nil {
		return nil, 0, err
	}
	for key, value := range c.apiHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := c.hc.Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		return nil, 0, &ErrNoPlankaAtURL{Reason: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return nil, 0, &ErrNoPlankaAtURL{Reason: fmt.Sprintf("status %d for /api/users/me", resp.StatusCode)}
	}

	me = &plankaUserResponse{}
	if resp.StatusCode == http.StatusOK {
		if err := migration.DecodeJSONLimited(resp.Body, me, maxResponseBytes); err != nil {
			return nil, 0, &ErrNoPlankaAtURL{Reason: err.Error()}
		}
	}
	return me, resp.StatusCode, nil
}

func (c *client) loginWithPassword(ctx context.Context, username, password string) error {
	body, err := json.Marshal(map[string]string{
		"emailOrUsername": username,
		"password":        password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/access-tokens", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.hc.Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		return &ErrNoPlankaAtURL{Reason: err.Error()}
	}
	defer resp.Body.Close()

	var result struct {
		Item string `json:"item"`
		Step string `json:"step"`
	}
	if err := migration.DecodeJSONLimited(resp.Body, &result, maxResponseBytes); err != nil && resp.StatusCode == http.StatusOK {
		return fmt.Errorf("could not decode planka login response: %w", err)
	}

	if result.Step != "" {
		return &ErrLoginStepRequired{Step: result.Step}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if result.Item == "" {
			return &ErrInvalidCredentials{}
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ErrInvalidCredentials{}
	default:
		return &ErrNoPlankaAtURL{Reason: fmt.Sprintf("status %d for /api/access-tokens", resp.StatusCode)}
	}

	c.apiKey, c.jwt, c.ownsJWT = "", result.Item, true

	me, status, err := c.probeMe(ctx)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return &ErrInvalidCredentials{}
	}
	return c.acceptUser(me)
}

// logout deletes the session created by loginWithPassword. Best effort.
func (c *client) logout(ctx context.Context) {
	if !c.ownsJWT || c.jwt == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/access-tokens/me", nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.jwt)
	resp, err := c.hc.Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		log.Debugf("[Planka Migration] Could not delete planka session: %s", err)
		return
	}
	resp.Body.Close()
	c.jwt = ""
}

func (c *client) apiHeaders() map[string]string {
	h := map[string]string{"Accept": "application/json"}
	if c.jwt != "" {
		h["Authorization"] = "Bearer " + c.jwt
	} else if c.apiKey != "" {
		h["X-Api-Key"] = c.apiKey
	}
	return h
}

// downloadHeaders: the /attachments/* routes only accept the accessToken cookie or the api key header, not Bearer.
func (c *client) downloadHeaders() http.Header {
	h := http.Header{}
	if c.jwt != "" {
		h.Set("Cookie", "accessToken="+c.jwt)
	} else if c.apiKey != "" {
		h.Set("X-Api-Key", c.apiKey)
	}
	return h
}

// getRaw performs an authenticated GET and decodes 2xx responses into out. It returns the status code for
// non-5xx responses so callers can act on 401/404.
func (c *client) getRaw(path string, query url.Values, out any) (int, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	resp, err := migration.DoGetWithClient(c.hc, u, c.apiHeaders())
	if err != nil {
		return 0, fmt.Errorf("planka request %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, nil
	}
	if err := migration.DecodeJSONLimited(resp.Body, out, maxResponseBytes); err != nil {
		return resp.StatusCode, fmt.Errorf("could not decode planka response for %s: %w", path, err)
	}
	return resp.StatusCode, nil
}

// get performs an authenticated GET and fails on any non-2xx status.
func (c *client) get(path string, query url.Values, out any) error {
	status, err := c.getRaw(path, query, out)
	if err != nil {
		return err
	}
	switch {
	case status == http.StatusUnauthorized:
		return &ErrInvalidCredentials{}
	case status < 200 || status >= 300:
		return fmt.Errorf("planka returned status %d for %s", status, path)
	}
	return nil
}

func maxAttachmentSize() int64 {
	return int64(config.GetMaxFileSizeInMBytes()) * int64(datasize.MB) //nolint:gosec // config value is small
}

// download fetches a file attachment, capped at files.maxsize.
func (c *client) download(attachmentID, filename string) (*bytes.Buffer, error) {
	u := c.baseURL + "/attachments/" + url.PathEscape(attachmentID) + "/download/" + url.PathEscape(filename)
	return migration.DownloadFileWithHeadersLimited(c.downloadHC, u, c.downloadHeaders(), maxAttachmentSize())
}
