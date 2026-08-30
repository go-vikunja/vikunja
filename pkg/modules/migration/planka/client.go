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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/utils"

	"github.com/c2h5oh/datasize"
)

// Budgets cap hostile response expansion and pagination (GHSA-wq92-8x3r-fm38).
const (
	defaultMaxResponseBytes = 4 << 20
	budgetResponseBytes     = 128 << 20
	budgetAttachmentBytes   = 64 << 20
	budgetEntities          = 50000
	// Count retries because hostile 5xx responses multiply logical requests.
	budgetRequestAttempts = 2000
)

const maxRedirects = 10

type jobBudget struct {
	maxResponseBytes   int64
	maxAggregateBytes  int64
	maxAttachmentBytes int64
	maxEntities        int64
	maxRequests        int64

	responseBytes   int64
	attachmentBytes int64
	entities        int64
	requests        int64
}

func newJobBudget() *jobBudget {
	return &jobBudget{
		maxResponseBytes:   defaultMaxResponseBytes,
		maxAggregateBytes:  budgetResponseBytes,
		maxAttachmentBytes: budgetAttachmentBytes,
		maxEntities:        budgetEntities,
		maxRequests:        budgetRequestAttempts,
	}
}

type countingReader struct {
	r io.Reader
	b *jobBudget
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.b.responseBytes += int64(n)
	return n, err
}

type attachmentCountingReadCloser struct {
	io.ReadCloser
	b *jobBudget
}

func (r *attachmentCountingReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.b.attachmentBytes += int64(n)
	return n, err
}

// budgetTransport marks budget failures as non-retryable.
type budgetTransport struct {
	base   http.RoundTripper
	budget *jobBudget
}

type attachmentBudgetTransport struct {
	base   http.RoundTripper
	budget *jobBudget
}

func (t *attachmentBudgetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if resp != nil {
		resp.Body = &attachmentCountingReadCloser{ReadCloser: resp.Body, b: t.budget}
	}
	return resp, err
}

func (t *budgetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.budget.requests >= t.budget.maxRequests {
		return nil, fmt.Errorf("%w: %w", utils.ErrDoNotRetry, &ErrImportBudgetExceeded{Exceeded: "outbound request attempts"})
	}
	t.budget.requests++
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

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
	budget     *jobBudget
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

	budget := newJobBudget()
	// Wrap the SSRF-safe transport so retries count as separate attempts.
	hc.Transport = &budgetTransport{base: hc.Transport, budget: budget}
	downloadHC.Transport = &attachmentBudgetTransport{
		base:   &budgetTransport{base: downloadHC.Transport, budget: budget},
		budget: budget,
	}

	return &client{
		baseURL:    u.String(),
		hc:         hc,
		downloadHC: downloadHC,
		budget:     budget,
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
		if err := c.decodeJSONLimited(resp.Body, me); err != nil {
			var budgetErr *ErrImportBudgetExceeded
			if errors.As(err, &budgetErr) {
				return nil, 0, err
			}
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
	if err := c.decodeJSONLimited(resp.Body, &result); err != nil {
		var budgetErr *ErrImportBudgetExceeded
		if errors.As(err, &budgetErr) {
			return err
		}
		if resp.StatusCode == http.StatusOK {
			return fmt.Errorf("could not decode planka login response: %w", err)
		}
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
	if err := c.decodeJSONLimited(resp.Body, out); err != nil {
		return resp.StatusCode, fmt.Errorf("could not decode planka response for %s: %w", path, err)
	}
	return resp.StatusCode, nil
}

func (c *client) decodeJSONLimited(body io.Reader, out any) error {
	limited := &io.LimitedReader{
		R: &countingReader{r: body, b: c.budget},
		N: c.budget.maxResponseBytes + 1,
	}
	raw, err := io.ReadAll(limited)
	if c.budget.responseBytes > c.budget.maxAggregateBytes {
		return &ErrImportBudgetExceeded{Exceeded: "aggregate response bytes"}
	}
	if int64(len(raw)) > c.budget.maxResponseBytes {
		return &ErrImportBudgetExceeded{Exceeded: "response bytes"}
	}
	if err != nil {
		return err
	}

	entityCount, err := countResponseEntities(raw, out)
	if err != nil {
		return err
	}
	if c.budget.entities+int64(entityCount) > c.budget.maxEntities {
		return &ErrImportBudgetExceeded{Exceeded: "decoded entities"}
	}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(out); err != nil {
		return err
	}
	c.budget.entities += int64(entityCount)
	return nil
}

type entityResponseKind uint8

const (
	entityResponseNone entityResponseKind = iota
	entityResponseProjects
	entityResponseBoard
	entityResponseListCards
	entityResponseComments
)

func countResponseEntities(raw []byte, out any) (int, error) {
	kind := entityKind(out)
	if kind == entityResponseNone {
		return 0, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return 0, err
	}
	return countEntityValues(decoder, token, kind, "")
}

func entityKind(out any) entityResponseKind {
	switch out.(type) {
	case *projectsResponse:
		return entityResponseProjects
	case *boardResponse:
		return entityResponseBoard
	case *listCardsResponse:
		return entityResponseListCards
	case *commentsResponse:
		return entityResponseComments
	default:
		return entityResponseNone
	}
}

func countEntityValues(decoder *json.Decoder, token json.Token, kind entityResponseKind, path string) (int, error) {
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return 0, nil
	}

	count := 0
	switch delim {
	case '{':
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return 0, err
			}
			value, err := decoder.Token()
			if err != nil {
				return 0, err
			}
			childPath := key.(string)
			if path != "" {
				childPath = path + "." + childPath
			}
			n, err := countEntityValues(decoder, value, kind, childPath)
			if err != nil {
				return 0, err
			}
			count += n
		}
	case '[':
		tracked := isEntityArray(kind, path)
		for decoder.More() {
			value, err := decoder.Token()
			if err != nil {
				return 0, err
			}
			if tracked {
				count++
			}
			n, err := countEntityValues(decoder, value, kind, path+"[]")
			if err != nil {
				return 0, err
			}
			count += n
		}
	}
	_, err := decoder.Token()
	return count, err
}

func isEntityArray(kind entityResponseKind, path string) bool {
	switch kind {
	case entityResponseNone:
		return false
	case entityResponseProjects:
		return path == "items" || path == "included.boards" ||
			path == "included.baseCustomFieldGroups" || path == "included.customFields"
	case entityResponseBoard:
		return isBoardEntityArray(path)
	case entityResponseListCards:
		return path == "items" || isBoardEntityArray(path)
	case entityResponseComments:
		return path == "items" || path == "included.users"
	default:
		return false
	}
}

func isBoardEntityArray(path string) bool {
	switch path {
	case "included.users", "included.labels", "included.lists", "included.cards",
		"included.cardLabels", "included.taskLists", "included.tasks", "included.attachments",
		"included.customFieldGroups", "included.customFields", "included.customFieldValues":
		return true
	default:
		return false
	}
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

func (c *client) download(attachmentID, filename string) (*bytes.Buffer, error) {
	u := c.baseURL + "/attachments/" + url.PathEscape(attachmentID) + "/download/" + url.PathEscape(filename)

	maxSize := maxAttachmentSize()
	limit := maxSize
	if remaining := c.budget.maxAttachmentBytes - c.budget.attachmentBytes; remaining < limit {
		limit = remaining
	}
	if limit <= 0 {
		return nil, &ErrImportBudgetExceeded{Exceeded: "attachment bytes"}
	}

	buf, err := migration.DownloadFileWithHeadersLimited(c.downloadHC, u, c.downloadHeaders(), limit)
	if err != nil {
		// A reduced limit means the job budget, not files.maxsize, was exhausted.
		var tooLarge files.ErrFileIsTooLarge
		if limit < maxSize && errors.As(err, &tooLarge) {
			return nil, &ErrImportBudgetExceeded{Exceeded: "attachment bytes"}
		}
		return nil, err
	}
	return buf, nil
}
