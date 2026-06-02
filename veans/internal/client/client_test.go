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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/veans/internal/output"
)

func TestMapHTTPError_StatusCodeMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		want   output.Code
	}{
		{"401 unauthorized -> auth", http.StatusUnauthorized, output.CodeAuth},
		{"403 forbidden -> auth", http.StatusForbidden, output.CodeAuth},
		{"404 not found -> not found", http.StatusNotFound, output.CodeNotFound},
		{"409 conflict -> conflict", http.StatusConflict, output.CodeConflict},
		{"429 too many requests -> rate limited", http.StatusTooManyRequests, output.CodeRateLimited},
		{"400 bad request -> validation", http.StatusBadRequest, output.CodeValidation},
		{"422 unprocessable -> validation", http.StatusUnprocessableEntity, output.CodeValidation},
		{"500 internal -> unknown", http.StatusInternalServerError, output.CodeUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := mapHTTPError("GET", "/foo", tc.status, []byte(`{"message":"boom"}`), 0)
			var oe *output.Error
			if !errors.As(err, &oe) {
				t.Fatalf("expected *output.Error, got %T", err)
			}
			if oe.Code != tc.want {
				t.Errorf("status %d: got code %q, want %q", tc.status, oe.Code, tc.want)
			}
		})
	}
}

func TestMapHTTPError_RetryAfterAppendedToMessage(t *testing.T) {
	retry := 7 * time.Second
	err := mapHTTPError("GET", "/foo", http.StatusTooManyRequests, []byte(`{"message":"slow down"}`), retry)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	if !strings.Contains(oe.Message, "retry-after") {
		t.Errorf("expected message to contain %q, got %q", "retry-after", oe.Message)
	}
	if !strings.Contains(oe.Message, retry.String()) {
		t.Errorf("expected message to contain duration %q, got %q", retry.String(), oe.Message)
	}
}

func TestMapHTTPError_BodyTruncation(t *testing.T) {
	// Build a > maxErrorMessageBytes (512) raw body that isn't valid JSON so
	// the message falls through to the raw-body branch.
	body := []byte(strings.Repeat("a", 600))
	err := mapHTTPError("GET", "/foo", http.StatusInternalServerError, body, 0)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	if !strings.HasSuffix(oe.Message, "…(truncated)") {
		t.Errorf("expected message to end with truncation marker, got %q", oe.Message)
	}
	if oe.Cause != nil {
		t.Errorf("expected Cause to be nil, got %v", oe.Cause)
	}
}

func TestMapHTTPError_VikunjaJSONTakesPrecedenceOverRawBody(t *testing.T) {
	body := []byte(`{"code":404,"message":"x"}`)
	err := mapHTTPError("GET", "/foo", http.StatusNotFound, body, 0)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	// The formatted message is "METHOD PATH: STATUS MSG"; assert it carries
	// the decoded message and not the raw JSON envelope.
	if !strings.HasSuffix(oe.Message, ": 404 x") {
		t.Errorf("expected formatted message to end with %q, got %q", ": 404 x", oe.Message)
	}
	if strings.Contains(oe.Message, `"code":404`) {
		t.Errorf("expected raw JSON body to be replaced by decoded message, got %q", oe.Message)
	}
}

func TestParseRetryAfter(t *testing.T) {
	future := time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)
	past := time.Now().Add(-30 * time.Second).UTC().Format(http.TimeFormat)

	cases := []struct {
		name string
		in   string
		want time.Duration
		// For HTTP-date inputs, the result is computed via time.Until; allow
		// a tolerance window.
		tolerance time.Duration
	}{
		{"empty", "", 0, 0},
		{"five seconds", "5", 5 * time.Second, 0},
		{"zero", "0", 0, 0},
		{"negative invalid", "-1", 0, 0},
		{"unparseable", "not a number", 0, 0},
		{"past http date", past, 0, 0},
		{"future http date ~30s", future, 30 * time.Second, 3 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseRetryAfter(tc.in)
			if tc.tolerance == 0 {
				if got != tc.want {
					t.Errorf("parseRetryAfter(%q) = %v, want %v", tc.in, got, tc.want)
				}
				return
			}
			diff := got - tc.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.tolerance {
				t.Errorf("parseRetryAfter(%q) = %v, want %v ± %v", tc.in, got, tc.want, tc.tolerance)
			}
		})
	}
}

func TestPaginationDone(t *testing.T) {
	cases := []struct {
		name       string
		page       int
		batchLen   int
		perPage    int
		totalPages int
		want       bool
	}{
		{"server says single page complete", 1, 50, 50, 1, true},
		{"server says more pages remain", 1, 50, 50, 2, false},
		{"server says we're on the last page", 2, 10, 50, 2, true},
		{"no header, full page -> not done", 1, 50, 50, 0, false},
		{"no header, short page -> done", 1, 10, 50, 0, true},
		{"no header, empty page -> done", 1, 0, 50, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := paginationDone(tc.page, tc.batchLen, tc.perPage, tc.totalPages)
			if got != tc.want {
				t.Errorf("paginationDone(page=%d, batch=%d, per=%d, total=%d) = %v, want %v",
					tc.page, tc.batchLen, tc.perPage, tc.totalPages, got, tc.want)
			}
		})
	}
}

func TestCreateBotUser_404TranslatesToBotUsersUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/user/bots" {
			http.Error(w, "unexpected route", http.StatusInternalServerError)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	_, err := c.CreateBotUser(context.Background(), "bot-test", "Test Bot")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T (%v)", err, err)
	}
	if oe.Code != output.CodeBotUsersUnavailable {
		t.Errorf("got code %q, want %q", oe.Code, output.CodeBotUsersUnavailable)
	}
}
