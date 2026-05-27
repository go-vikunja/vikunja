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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/veans/internal/output"
)

// Client is a thin JSON wrapper around the Vikunja REST API. It holds the
// server base URL and a bearer token (either a JWT from POST /login or an
// API token minted via PUT /tokens). Every method in this package is a thin
// shim over Do.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// UserAgent is the value sent in the User-Agent header on every request.
// main sets this at startup with the linker-injected version + the
// runtime os/arch (e.g. "veans/0.3.1 (linux/amd64)"). Tests get the
// default "veans/dev". Vikunja admins see this in their access logs.
var UserAgent = "veans/dev"

// defaultHTTPTimeout is the timeout applied to the HTTP client returned by
// New. Callers that need a different value (e.g. the runtime loader honoring
// `http_timeout` from .veans.yml) can overwrite HTTPClient.Timeout after
// construction.
const defaultHTTPTimeout = 30 * time.Second

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// vikunjaError matches `web.HTTPError` on the wire.
type vikunjaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Do performs a single JSON request against /api/v1<path>. body, if non-nil,
// is JSON-marshalled. out, if non-nil, is JSON-unmarshalled. query is appended
// as URL-encoded params.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	full := c.BaseURL + "/api/v1" + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, full, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return output.Wrap(output.CodeUnknown, err, "%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return mapHTTPError(method, path, resp.StatusCode, respBody,
			parseRetryAfter(resp.Header.Get("Retry-After")))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode %s %s: %w", method, path, err)
		}
	}
	return nil
}

// DoPaginated is like Do but also returns the total page count parsed from
// the `x-pagination-total-pages` response header (0 if the header is
// missing or unparseable). Used by the list endpoints so paging terminates
// against the authoritative server count, not a `len(batch) < per_page`
// heuristic that loops one extra time on exact-multiple totals.
func (c *Client) DoPaginated(ctx context.Context, method, path string, query url.Values, out any) (totalPages int, err error) {
	full := c.BaseURL + "/api/v1" + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, full, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, output.Wrap(output.CodeUnknown, err, "%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return 0, mapHTTPError(method, path, resp.StatusCode, respBody,
			parseRetryAfter(resp.Header.Get("Retry-After")))
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return 0, fmt.Errorf("decode %s %s: %w", method, path, err)
		}
	}
	if v := resp.Header.Get("x-pagination-total-pages"); v != "" {
		if n, perr := strconv.Atoi(v); perr == nil {
			totalPages = n
		}
	}
	return totalPages, nil
}

// DoRaw is the escape hatch used by `veans api`. It returns the raw response
// body, status, and the parsed Retry-After (if any). Auth + base URL handling
// matches Do. The caller is responsible for deciding whether to surface the
// body to stdout — non-2xx bodies should NOT be written there (the contract is
// "stdout is for the success payload; errors go through the envelope on
// stderr"); see commands/api.go for the canonical handling.
func (c *Client) DoRaw(ctx context.Context, method, path string, query url.Values, body []byte) (status int, respBody []byte, retryAfter time.Duration, err error) {
	full := c.BaseURL + "/api/v1" + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, full, br)
	if err != nil {
		return 0, nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	return resp.StatusCode, respBody, parseRetryAfter(resp.Header.Get("Retry-After")), err
}

// paginationDone reports whether a paged GET has consumed every page,
// preferring the server's x-pagination-total-pages count when present and
// falling back to the len(batch) < per_page heuristic when the header is
// missing (older server / proxy stripped). Centralized so all list
// endpoints terminate identically.
func paginationDone(page, batchLen, perPage, totalPages int) bool {
	if totalPages > 0 {
		return page >= totalPages
	}
	return batchLen < perPage
}

// maxBodyBytes caps the size of any response body we'll read into memory.
// Vikunja JSON payloads are far smaller; the cap exists so a misbehaving
// proxy can't OOM the CLI by streaming an unbounded body.
const maxBodyBytes = 32 * 1024 * 1024 // 32 MiB

// parseRetryAfter parses an HTTP Retry-After header value. Supports both
// the delta-seconds form and the HTTP-date form; returns 0 on unparseable
// or empty input.
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func mapHTTPError(method, path string, status int, body []byte, retryAfter time.Duration) error {
	var ve vikunjaError
	_ = json.Unmarshal(body, &ve)
	msg := strings.TrimSpace(ve.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(status)
		}
	}
	// Truncate so an HTML error page (e.g. from a reverse proxy) doesn't
	// dump several KB into the agent's stderr envelope.
	if len(msg) > maxErrorMessageBytes {
		msg = msg[:maxErrorMessageBytes] + "…(truncated)"
	}

	var code output.Code
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		code = output.CodeAuth
	case status == http.StatusNotFound:
		code = output.CodeNotFound
	case status == http.StatusConflict:
		code = output.CodeConflict
	case status == http.StatusTooManyRequests:
		code = output.CodeRateLimited
	case status >= 400 && status < 500:
		code = output.CodeValidation
	default:
		code = output.CodeUnknown
	}

	formatted := fmt.Sprintf("%s %s: %d %s", method, path, status, msg)
	if retryAfter > 0 {
		formatted += fmt.Sprintf(" (retry-after %s)", retryAfter)
	}
	return &output.Error{
		Code:    code,
		Message: formatted,
	}
}

// maxErrorMessageBytes caps how much upstream-error text we embed in the
// envelope's `error` field. Anything longer is almost always an HTML page
// from a proxy and useless for the agent to read.
const maxErrorMessageBytes = 512
