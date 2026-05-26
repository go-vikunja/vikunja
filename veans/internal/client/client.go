package client

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
	UserAgent  string
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		UserAgent:  "veans/0.1",
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
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return output.Wrap(output.CodeUnknown, err, "%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return mapHTTPError(method, path, resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode %s %s: %w", method, path, err)
		}
	}
	return nil
}

// DoRaw is the escape hatch used by `veans api`. It returns the raw response
// body and status. Auth + base URL handling matches Do.
func (c *Client) DoRaw(ctx context.Context, method, path string, query url.Values, body []byte) (status int, respBody []byte, err error) {
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
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, err = io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, err
}

func mapHTTPError(method, path string, status int, body []byte) error {
	var ve vikunjaError
	_ = json.Unmarshal(body, &ve)
	msg := strings.TrimSpace(ve.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(status)
		}
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

	return &output.Error{
		Code:    code,
		Message: fmt.Sprintf("%s %s: %d %s", method, path, status, msg),
		Cause:   errors.New(msg),
	}
}
