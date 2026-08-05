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
			err := mapHTTPError("GET", "/foo", tc.status, []byte(`{"detail":"boom"}`), 0)
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
	err := mapHTTPError("GET", "/foo", http.StatusTooManyRequests, []byte(`{"detail":"slow down"}`), retry)
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

func TestMapHTTPError_VikunjaProblemJSONTakesPrecedenceOverRawBody(t *testing.T) {
	// v2 returns RFC 9457 problem+json: the message is in `detail`, and `code`
	// carries Vikunja's numeric domain error code (not the HTTP status).
	body := []byte(`{"status":404,"title":"Not Found","detail":"x","code":3001}`)
	err := mapHTTPError("GET", "/foo", http.StatusNotFound, body, 0)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	// The formatted message is "METHOD PATH: STATUS MSG"; assert it carries
	// the decoded `detail` and not the raw JSON envelope.
	if !strings.HasSuffix(oe.Message, ": 404 x") {
		t.Errorf("expected formatted message to end with %q, got %q", ": 404 x", oe.Message)
	}
	if strings.Contains(oe.Message, `"code"`) {
		t.Errorf("expected raw JSON body to be replaced by decoded detail, got %q", oe.Message)
	}
}

func TestMapHTTPError_FallsBackToTitleWhenNoDetail(t *testing.T) {
	// A problem+json body with no `detail` (e.g. Huma's own schema-validation
	// 422 sometimes only sets title) falls back to `title`.
	body := []byte(`{"status":422,"title":"Unprocessable Entity"}`)
	err := mapHTTPError("PATCH", "/tasks/1", http.StatusUnprocessableEntity, body, 0)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	if !strings.HasSuffix(oe.Message, ": 422 Unprocessable Entity") {
		t.Errorf("expected title fallback, got %q", oe.Message)
	}
}

func TestMapHTTPError_FallsBackToLegacyMessage(t *testing.T) {
	// Defensive: a stray legacy/proxy body with only v1's `message` field
	// still yields the message text rather than the raw JSON.
	body := []byte(`{"code":403,"message":"forbidden"}`)
	err := mapHTTPError("GET", "/foo", http.StatusForbidden, body, 0)
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *output.Error, got %T", err)
	}
	if !strings.HasSuffix(oe.Message, ": 403 forbidden") {
		t.Errorf("expected legacy message fallback, got %q", oe.Message)
	}
	if strings.Contains(oe.Message, `"message"`) {
		t.Errorf("expected raw JSON to be replaced by the message text, got %q", oe.Message)
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

func TestCreateBotUser_404TranslatesToBotUsersUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/user/bots" {
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

// TestListProjects_PaginatesEnvelope verifies the v2 list shape: each page is
// the {items,total,page,per_page,total_pages} envelope, and ListProjects keeps
// requesting until page >= total_pages, accumulating every item.
func TestListProjects_PaginatesEnvelope(t *testing.T) {
	var gotPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/projects" {
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusInternalServerError)
			return
		}
		page := r.URL.Query().Get("page")
		gotPages = append(gotPages, page)
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"title":"a"},{"id":2,"title":"b"}],"total":3,"page":1,"per_page":2,"total_pages":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":3,"title":"c"}],"total":3,"page":2,"per_page":2,"total_pages":2}`))
		default:
			t.Errorf("unexpected page %q (would loop past the end)", page)
			http.Error(w, "no such page", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	projects, err := New(srv.URL, "tk").ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects accumulated across 2 pages, got %d", len(projects))
	}
	if len(gotPages) != 2 || gotPages[0] != "1" || gotPages[1] != "2" {
		t.Fatalf("expected exactly pages [1 2], got %v", gotPages)
	}
}

// TestListTaskComments_PaginatesEnvelope guards the truncation bug: the v2
// comments endpoint is server-paginated, so a task with >50 comments spans
// multiple pages and ListTaskComments must accumulate them all, not stop at
// page 1.
func TestListTaskComments_PaginatesEnvelope(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tasks/9/comments" {
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusInternalServerError)
			return
		}
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"comment":"a"},{"id":2,"comment":"b"}],"total":3,"page":1,"per_page":2,"total_pages":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":3,"comment":"c"}],"total":3,"page":2,"per_page":2,"total_pages":2}`))
		default:
			t.Errorf("unexpected page %q", page)
			http.Error(w, "no such page", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	comments, err := New(srv.URL, "tk").ListTaskComments(context.Background(), 9)
	if err != nil {
		t.Fatalf("ListTaskComments: %v", err)
	}
	if len(comments) != 3 {
		t.Fatalf("expected 3 comments across 2 pages, got %d (truncation regression?)", len(comments))
	}
	if len(pages) != 2 {
		t.Fatalf("expected to fetch 2 pages, got %v", pages)
	}
}

// TestListBuckets_SingleFetchDoesNotPage pins the opposite invariant: the
// buckets model returns every row in one page, so ListBuckets must issue a
// single request even when the envelope's total_pages is >1 — paging would
// re-fetch and duplicate the buckets.
func TestListBuckets_SingleFetchDoesNotPage(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		if requests > 1 {
			t.Errorf("ListBuckets paged a single-page endpoint (request %d) — would duplicate", requests)
		}
		w.Header().Set("Content-Type", "application/json")
		// total_pages deliberately > 1 to prove ListBuckets ignores it.
		_, _ = w.Write([]byte(`{"items":[{"id":1,"title":"Todo"},{"id":2,"title":"Doing"}],"total":2,"page":1,"per_page":1,"total_pages":2}`))
	}))
	defer srv.Close()

	buckets, err := New(srv.URL, "tk").ListBuckets(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected exactly 1 request, got %d", requests)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected the 2 buckets from the single page, got %d", len(buckets))
	}
}

// TestListProjectViews_UnwrapsEnvelope pins that a previously-single-GET list
// (project views) now unwraps .items from the standard list envelope.
func TestListProjectViews_UnwrapsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/projects/7/views" {
			http.Error(w, "unexpected path "+r.URL.Path, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":10,"title":"Kanban","view_kind":"kanban"}],"total":1,"page":1,"per_page":50,"total_pages":1}`))
	}))
	defer srv.Close()

	views, err := New(srv.URL, "tk").ListProjectViews(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListProjectViews: %v", err)
	}
	if len(views) != 1 || views[0].ViewKind != ViewKindKanban {
		t.Fatalf("expected one kanban view unwrapped from .items, got %+v", views)
	}
}
