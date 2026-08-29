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

package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
	"code.vikunja.io/veans/internal/status"
)

func TestValidateBotUsername(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantValid bool
	}{
		// Valid names.
		{"valid simple", "bot-foo", true},
		{"valid multi-hyphen", "bot-foo-bar", true},
		{"valid digits", "bot-foo123", true},
		{"valid underscore", "bot-foo_bar", true},
		{"valid dot", "bot-foo.bar", true},
		{"valid single letter", "bot-a", true},

		// Invalid: missing/malformed bot- prefix.
		{"missing prefix", "foo", false},
		{"uppercase prefix", "Bot-foo", false},
		{"empty", "", false},

		// Invalid: forbidden characters in the body.
		{"space after prefix", "bot- foo", false},
		{"comma", "bot-foo,bar", false},
		{"uppercase body", "bot-FOO", false},
		{"bang", "bot-foo!", false},
		{"space in body", "bot-foo bar", false},

		// Invalid: reserved link-share pattern.
		{"link-share-0", "bot-link-share-0", false},
		{"link-share-1", "bot-link-share-1", false},
		{"link-share-42", "bot-link-share-42", false},

		// Edge: regex `^bot-[a-z0-9][a-z0-9._-]*$` requires at least one char
		// after the `bot-` prefix, so a bare prefix is rejected.
		{"bare prefix", "bot-", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBotUsername(tc.input)
			if tc.wantValid {
				if err != nil {
					t.Fatalf("validateBotUsername(%q) = %v, want nil", tc.input, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateBotUsername(%q) = nil, want error", tc.input)
			}
			var oe *output.Error
			if !errors.As(err, &oe) {
				t.Fatalf("validateBotUsername(%q): expected *output.Error, got %T", tc.input, err)
			}
			if oe.Code != output.CodeValidation {
				t.Errorf("validateBotUsername(%q): code = %q, want %q", tc.input, oe.Code, output.CodeValidation)
			}
		})
	}
}

// queuePrompter is a richer scriptedPrompter that can also return an error
// on a chosen call (to simulate stdin read failures) and tracks how many
// times ReadLine was invoked. Defined locally because the existing
// scriptedPrompter in botuser_test.go can't inject errors.
type queuePrompter struct {
	answers []string
	err     error // returned on every call once exhausted, or immediately if no answers
	calls   int
}

func (q *queuePrompter) ReadLine(_ string) (string, error) {
	q.calls++
	if q.err != nil {
		return "", q.err
	}
	if q.calls-1 >= len(q.answers) {
		return "", nil
	}
	return q.answers[q.calls-1], nil
}

func (q *queuePrompter) ReadPassword(_ string) (string, error) { return "", nil }

// errReadFailure is a sentinel used to simulate a stdin read failure
// inside the prompter. Kept at package level to satisfy err113's
// preference for static errors (test files are exempt, but using a
// named value reads more clearly than fmt.Errorf at the call site).
var errReadFailure = errors.New("simulated stdin read failure")

func TestConfirmOverwriteExistingConfig(t *testing.T) {
	t.Run("file missing — no prompt", func(t *testing.T) {
		dir := t.TempDir()
		p := &queuePrompter{}
		opts := &Options{ConfigPath: filepath.Join(dir, "does-not-exist.yml")}
		if err := confirmOverwriteExistingConfig(opts, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.calls != 0 {
			t.Errorf("prompter called %d times, want 0", p.calls)
		}
	})

	t.Run("OverwriteExistingConfig=true — no prompt", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
			t.Fatal(err)
		}
		p := &queuePrompter{}
		opts := &Options{ConfigPath: path, OverwriteExistingConfig: true}
		if err := confirmOverwriteExistingConfig(opts, p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.calls != 0 {
			t.Errorf("prompter called %d times, want 0", p.calls)
		}
	})

	t.Run("answers — yes/no/error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
			t.Fatal(err)
		}

		yesAnswers := []string{"y", "yes", "Y", "Yes", "  yes  "}
		for _, ans := range yesAnswers {
			p := &queuePrompter{answers: []string{ans}}
			opts := &Options{ConfigPath: path}
			if err := confirmOverwriteExistingConfig(opts, p); err != nil {
				t.Errorf("answer %q: unexpected error: %v", ans, err)
			}
		}

		// "n", "", and any other input → conflict.
		noAnswers := []string{"n", "", "no", "garbage"}
		for _, ans := range noAnswers {
			p := &queuePrompter{answers: []string{ans}}
			opts := &Options{ConfigPath: path}
			err := confirmOverwriteExistingConfig(opts, p)
			if err == nil {
				t.Errorf("answer %q: expected error, got nil", ans)
				continue
			}
			var oe *output.Error
			if !errors.As(err, &oe) {
				t.Errorf("answer %q: want *output.Error, got %T", ans, err)
				continue
			}
			if oe.Code != output.CodeConflict {
				t.Errorf("answer %q: code = %q, want %q", ans, oe.Code, output.CodeConflict)
			}
			if !strings.Contains(oe.Message, path) {
				t.Errorf("answer %q: message %q should contain config path %q", ans, oe.Message, path)
			}
		}

		// Prompter read failure → wrapped as CodeUnknown.
		p := &queuePrompter{err: errReadFailure}
		opts := &Options{ConfigPath: path}
		err := confirmOverwriteExistingConfig(opts, p)
		if err == nil {
			t.Fatal("expected error from prompter failure, got nil")
		}
		var oe *output.Error
		if !errors.As(err, &oe) {
			t.Fatalf("want *output.Error, got %T", err)
		}
		if oe.Code != output.CodeUnknown {
			t.Errorf("code = %q, want %q", oe.Code, output.CodeUnknown)
		}
		if !errors.Is(err, errReadFailure) {
			t.Errorf("wrapped error should unwrap to errReadFailure, got %v", err)
		}
	})
}

// bucketServer is a minimal httptest server modelling
// GET/POST /api/v2/projects/{p}/views/{v}/buckets. The caller pre-seeds
// existing buckets; POST requests append to that list with a synthetic ID.
// GET returns the standard v2 list envelope; POST returns the bare bucket.
type bucketServer struct {
	mu       sync.Mutex
	existing []*client.Bucket
	creates  []*client.Bucket // recorded create payloads (in order)
	nextID   int64
}

func newBucketServer(seed []*client.Bucket) *bucketServer {
	s := &bucketServer{existing: seed, nextID: 1000}
	for _, b := range seed {
		if b.ID >= s.nextID {
			s.nextID = b.ID + 1
		}
	}
	return s
}

func (s *bucketServer) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path is /api/v2/projects/{p}/views/{v}/buckets.
		if !strings.HasSuffix(r.URL.Path, "/buckets") || !strings.Contains(r.URL.Path, "/views/") {
			http.Error(w, "unexpected path: "+r.URL.Path, http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			// v2 list envelope; the buckets list isn't server-paginated.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items":       s.existing,
				"total":       len(s.existing),
				"page":        1,
				"per_page":    50,
				"total_pages": 1,
			})
		case http.MethodPost:
			var b client.Bucket
			if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			b.ID = s.nextID
			s.nextID++
			created := &client.Bucket{ID: b.ID, Title: b.Title, ProjectViewID: b.ProjectViewID}
			s.existing = append(s.existing, created)
			s.creates = append(s.creates, created)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed: "+r.Method, http.StatusMethodNotAllowed)
		}
	})
}

// allBucketsSeed returns one bucket per canonical status (using the
// canonical title from BucketTitle).
func allBucketsSeed() []*client.Bucket {
	var out []*client.Bucket
	id := int64(10)
	for _, s := range status.All() {
		out = append(out, &client.Bucket{ID: id, Title: s.BucketTitle()})
		id++
	}
	return out
}

func TestBootstrapBuckets_AllPresent_NoPrompt(t *testing.T) {
	srv := newBucketServer(allBucketsSeed())
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	p := &queuePrompter{} // any call would still return "" but we'll assert calls==0
	var buf bytes.Buffer
	opts := &Options{Out: &buf}

	buckets, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.calls != 0 {
		t.Errorf("prompter called %d times, want 0 (no missing buckets means no prompt)", p.calls)
	}
	if len(srv.creates) != 0 {
		t.Errorf("CreateBucket called %d times, want 0", len(srv.creates))
	}
	if buckets.Todo == 0 || buckets.InProgress == 0 || buckets.InReview == 0 || buckets.Done == 0 || buckets.Scrapped == 0 {
		t.Errorf("expected all bucket IDs populated, got %+v", buckets)
	}
}

func TestBootstrapBuckets_AutoApprove_CreatesMissing(t *testing.T) {
	// Seed only Todo; the other four are missing.
	srv := newBucketServer([]*client.Bucket{
		{ID: 10, Title: status.Todo.BucketTitle()},
	})
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	p := &queuePrompter{}
	var buf bytes.Buffer
	opts := &Options{Out: &buf, AutoApproveBuckets: true}

	buckets, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.calls != 0 {
		t.Errorf("AutoApprove should skip prompt; got %d calls", p.calls)
	}
	if len(srv.creates) != 4 {
		t.Errorf("expected 4 buckets created (the missing ones), got %d", len(srv.creates))
	}
	if buckets.Todo != 10 {
		t.Errorf("existing Todo bucket should be reused, got id=%d", buckets.Todo)
	}
	if buckets.InProgress == 0 || buckets.InReview == 0 || buckets.Done == 0 || buckets.Scrapped == 0 {
		t.Errorf("missing buckets not populated: %+v", buckets)
	}
}

func TestBootstrapBuckets_PromptDeclined(t *testing.T) {
	srv := newBucketServer([]*client.Bucket{
		{ID: 10, Title: status.Todo.BucketTitle()},
	})
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	p := &queuePrompter{answers: []string{"n"}}
	var buf bytes.Buffer
	opts := &Options{Out: &buf}

	_, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err == nil {
		t.Fatal("expected error on declined prompt, got nil")
	}
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("want *output.Error, got %T", err)
	}
	if oe.Code != output.CodeValidation {
		t.Errorf("code = %q, want %q", oe.Code, output.CodeValidation)
	}
	// Message should mention at least one of the missing canonical titles.
	mentionsMissing := false
	for _, s := range status.All() {
		if s == status.Todo {
			continue // not missing
		}
		if strings.Contains(oe.Message, s.BucketTitle()) {
			mentionsMissing = true
			break
		}
	}
	if !mentionsMissing {
		t.Errorf("error message %q should mention missing bucket titles", oe.Message)
	}
	if len(srv.creates) != 0 {
		t.Errorf("no buckets should be created on decline; got %d", len(srv.creates))
	}
}

func TestBootstrapBuckets_PromptAborted(t *testing.T) {
	srv := newBucketServer([]*client.Bucket{
		{ID: 10, Title: status.Todo.BucketTitle()},
	})
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	// Five garbage answers (each within the unknown limit), then "a" → abort.
	p := &queuePrompter{answers: []string{"huh", "what", "?", "??", "???", "a"}}
	var buf bytes.Buffer
	opts := &Options{Out: &buf}

	_, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err == nil {
		t.Fatal("expected abort error, got nil")
	}
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("want *output.Error, got %T", err)
	}
	if oe.Code != output.CodeValidation {
		t.Errorf("code = %q, want %q", oe.Code, output.CodeValidation)
	}
	if !strings.Contains(oe.Message, "abort") {
		t.Errorf("message %q should mention user abort", oe.Message)
	}
	if len(srv.creates) != 0 {
		t.Errorf("no buckets should be created on abort; got %d", len(srv.creates))
	}
}

func TestBootstrapBuckets_PromptUnknownCap(t *testing.T) {
	srv := newBucketServer([]*client.Bucket{
		{ID: 10, Title: status.Todo.BucketTitle()},
	})
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	// Six garbage answers — exceeds maxUnknownAnswers (5).
	p := &queuePrompter{answers: []string{"huh", "what", "?", "??", "???", "still no"}}
	var buf bytes.Buffer
	opts := &Options{Out: &buf}

	_, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err == nil {
		t.Fatal("expected cap error, got nil")
	}
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("want *output.Error, got %T", err)
	}
	if oe.Code != output.CodeValidation {
		t.Errorf("code = %q, want %q", oe.Code, output.CodeValidation)
	}
	if !strings.Contains(oe.Message, fmt.Sprintf("%d attempts", 5)) {
		t.Errorf("message %q should mention 5 attempts", oe.Message)
	}
}

func TestBootstrapBuckets_PromptAccepted(t *testing.T) {
	srv := newBucketServer([]*client.Bucket{
		{ID: 10, Title: status.Todo.BucketTitle()},
	})
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	c := client.New(ts.URL, "token")
	p := &queuePrompter{answers: []string{"y"}}
	var buf bytes.Buffer
	opts := &Options{Out: &buf}

	buckets, err := bootstrapBuckets(context.Background(), c, 1, 2, opts, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(srv.creates) != 4 {
		t.Errorf("expected 4 missing buckets created, got %d", len(srv.creates))
	}
	if buckets.Todo != 10 || buckets.InProgress == 0 || buckets.InReview == 0 || buckets.Done == 0 || buckets.Scrapped == 0 {
		t.Errorf("buckets not populated: %+v", buckets)
	}
}
