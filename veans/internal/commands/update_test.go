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

package commands

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
)

func TestComposeDescription_FullReplace(t *testing.T) {
	f := &updateFlags{description: "new body", descriptionIsSet: true}
	got, changed, err := composeDescription("old body", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "new body" {
		t.Fatalf("got %q changed=%v", got, changed)
	}
}

func TestComposeDescription_SurgicalReplace(t *testing.T) {
	f := &updateFlags{
		replaceOld: "TODO",
		replaceNew: "DONE",
	}
	got, changed, err := composeDescription("- [ ] TODO part 1\n- [ ] something else", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !strings.Contains(got, "DONE part 1") {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_ReplaceNotUnique(t *testing.T) {
	f := &updateFlags{
		replaceOld: "x",
		replaceNew: "y",
	}
	if _, _, err := composeDescription("xxx", f); err == nil {
		t.Fatal("expected error on non-unique match")
	}
}

func TestComposeDescription_ReplaceNotFound(t *testing.T) {
	f := &updateFlags{
		replaceOld: "missing",
		replaceNew: "y",
	}
	if _, _, err := composeDescription("hello", f); err == nil {
		t.Fatal("expected error on no match")
	}
}

func TestComposeDescription_Append(t *testing.T) {
	f := &updateFlags{descriptionApp: "## Notes"}
	got, changed, err := composeDescription("body", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "body\n## Notes" {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_AppendOnEmpty(t *testing.T) {
	f := &updateFlags{descriptionApp: "first line"}
	got, changed, err := composeDescription("", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "first line" {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_NoOp(t *testing.T) {
	f := &updateFlags{}
	got, changed, err := composeDescription("body", f)
	if err != nil {
		t.Fatal(err)
	}
	if changed || got != "body" {
		t.Fatalf("expected no-op, got %q changed=%v", got, changed)
	}
}

func TestComposeDescription_ReplaceNewWithoutOld(t *testing.T) {
	f := &updateFlags{replaceNew: "y"}
	if _, _, err := composeDescription("body", f); err == nil {
		t.Fatal("expected error: --description-replace-new without --description-replace-old")
	}
}

func TestNormalizeLabelTitle(t *testing.T) {
	cases := map[string]string{
		"foo":                    "veans:foo",
		"veans:bar":              "veans:bar",
		"  baz  ":                "veans:baz",
		"veans:already-prefixed": "veans:already-prefixed",
	}
	for in, want := range cases {
		if got := normalizeLabelTitle(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

// recordedCall captures one HTTP request made during a runUpdate invocation.
// The fake server appends these in order; tests assert against the sequence.
type recordedCall struct {
	method string
	path   string
}

// startRecordingServer spins up an httptest.Server that answers every
// Vikunja endpoint runUpdate touches with the minimum payload needed to
// keep the call chain alive, while appending each (method, path) to the
// returned slice. The server intentionally does NOT validate request
// bodies — the goal here is to pin call ORDER, not wire shape (which the
// e2e suite already covers).
func startRecordingServer(t *testing.T) (*httptest.Server, *[]recordedCall) {
	t.Helper()
	var (
		mu    sync.Mutex
		calls []recordedCall
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, recordedCall{method: r.Method, path: r.URL.Path})
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tasks/42":
			// Initial fetch + the final refetch both land here. Return a
			// fixed task with an empty label set — labels.go's
			// findLabelOnTask only iterates t.Labels.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 42, "title": "t", "updated": "2026-01-01T00:00:00Z",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/tasks/42/comments":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "comment": ""})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/tasks/42":
			// UpdateTask (merge-patch). Echo back the id so the encoder
			// downstream is happy with a non-nil Task.
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/projects/") && strings.HasSuffix(r.URL.Path, "/tasks"):
			// Bucket-task move (PUT .../buckets/{b}/tasks).
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/labels":
			// getOrCreateLabelByTitle's lookup. Empty envelope → falls through
			// to label creation.
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "total_pages": 1})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/labels":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99, "title": "veans:bug"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/tasks/42/labels":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

// newTestRuntime returns a *runtime suitable for runUpdate tests. The
// bucket IDs match the order the canonical statuses appear in
// status.BucketTitleAliases — Todo=10, InProgress=11, etc. — so test
// assertions can name the moved-to bucket by id.
func newTestRuntime(serverURL string) *runtime {
	return &runtime{
		cfg: &config.Config{
			Server:    serverURL,
			ProjectID: 7,
			ViewID:    1,
			Buckets: config.Buckets{
				Todo: 10, InProgress: 11, InReview: 12, Done: 13, Scrapped: 14,
			},
		},
		client: client.New(serverURL, "tk"),
	}
}

// TestRunUpdate_ScrappedOrdersCommentUpdateMove pins the audit-trail
// ordering invariant from CLAUDE.md ("Comments for `--status scrapped`
// post BEFORE the bucket move so the audit trail reads in chronological
// order"). A refactor that hoists the bucket move ahead of the comment
// would silently swap the timeline; this test fails if that happens.
func TestRunUpdate_ScrappedOrdersCommentUpdateMove(t *testing.T) {
	srv, calls := startRecordingServer(t)
	rt := newTestRuntime(srv.URL)

	if _, err := runUpdate(context.Background(), rt, 42, &updateFlags{
		statusName: "scrapped",
		reason:     "obsolete",
	}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	want := []recordedCall{
		{http.MethodGet, "/api/v2/tasks/42"},                            // current task fetch
		{http.MethodPost, "/api/v2/tasks/42/comments"},                  // "Scrapped: obsolete"
		{http.MethodPatch, "/api/v2/tasks/42"},                          // field update (done=true)
		{http.MethodPut, "/api/v2/projects/7/views/1/buckets/14/tasks"}, // bucket move to Scrapped
		{http.MethodGet, "/api/v2/tasks/42"},                            // refetch with new bucket
	}
	if !reflect.DeepEqual(*calls, want) {
		t.Fatalf("call order mismatch:\nwant: %#v\ngot:  %#v", want, *calls)
	}
}

// TestRunUpdate_BucketMoveBeforeLabelAdd pins the second ordering
// invariant from CLAUDE.md ("MoveTaskToBucket runs AFTER the field
// update so a status transition can't clobber freshly attached labels").
// Equivalently — and what this test asserts — labels are attached AFTER
// the bucket move, so the post-move state is the one we then refetch.
// A refactor that consolidates "all bucket-related work" by moving the
// labels-add loop ahead of MoveTaskToBucket would compile and silently
// regress label visibility; this test catches that.
func TestRunUpdate_BucketMoveBeforeLabelAdd(t *testing.T) {
	srv, calls := startRecordingServer(t)
	rt := newTestRuntime(srv.URL)

	if _, err := runUpdate(context.Background(), rt, 42, &updateFlags{
		statusName: "in-progress",
		addLabels:  []string{"bug"},
	}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	want := []recordedCall{
		{http.MethodGet, "/api/v2/tasks/42"},                            // current task fetch
		{http.MethodPatch, "/api/v2/tasks/42"},                          // field update (done=false)
		{http.MethodPut, "/api/v2/projects/7/views/1/buckets/11/tasks"}, // bucket move to In Progress
		{http.MethodGet, "/api/v2/labels"},                              // getOrCreateLabelByTitle lookup
		{http.MethodPost, "/api/v2/labels"},                             // create veans:bug
		{http.MethodPost, "/api/v2/tasks/42/labels"},                    // attach label
		{http.MethodGet, "/api/v2/tasks/42"},                            // refetch
	}
	if !reflect.DeepEqual(*calls, want) {
		t.Fatalf("call order mismatch:\nwant: %#v\ngot:  %#v", want, *calls)
	}
}

// startPatchCapturingServer answers the endpoints a field-only runUpdate
// touches (no labels) and records the raw JSON body of the merge-patch
// PATCH /tasks/{id} request so a test can assert exactly which fields were
// sent. The seeded task carries a description and priority that a partial
// update must leave untouched (issue #2962).
func startPatchCapturingServer(t *testing.T) (*httptest.Server, *[]byte) {
	t.Helper()
	var (
		mu        sync.Mutex
		patchBody []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tasks/42":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 42, "title": "t", "description": "keep me",
				"priority": 4, "updated": "2026-01-01T00:00:00Z",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/tasks/42":
			b, _ := io.ReadAll(r.Body)
			mu.Lock()
			patchBody = b
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/projects/") && strings.HasSuffix(r.URL.Path, "/tasks"):
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &patchBody
}

// TestRunUpdate_TitleOnlyPatchOmitsOtherFields is the #2962 acceptance test for
// a title-only update: the merge-patch body must contain ONLY title, so the
// stored description/priority/done are left intact server-side.
func TestRunUpdate_TitleOnlyPatchOmitsOtherFields(t *testing.T) {
	srv, patchBody := startPatchCapturingServer(t)
	rt := newTestRuntime(srv.URL)

	if _, err := runUpdate(context.Background(), rt, 42, &updateFlags{title: "new title"}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	body := decodePatchBody(t, *patchBody)
	if body["title"] != "new title" {
		t.Errorf("patch should set title; got %v", body["title"])
	}
	for _, k := range []string{"description", "priority", "done"} {
		if _, present := body[k]; present {
			t.Errorf("title-only update must not send %q (merge-patch would clobber it); body=%s", k, *patchBody)
		}
	}
}

// TestRunUpdate_StatusOnlyPatchPreservesDescriptionAndPriority is the #2962
// acceptance test for a status-only update: the merge-patch body carries the
// done flag (and nothing else), so description and priority survive — the
// regression the whole-object POST caused.
func TestRunUpdate_StatusOnlyPatchPreservesDescriptionAndPriority(t *testing.T) {
	srv, patchBody := startPatchCapturingServer(t)
	rt := newTestRuntime(srv.URL)

	if _, err := runUpdate(context.Background(), rt, 42, &updateFlags{statusName: "in-progress"}); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	body := decodePatchBody(t, *patchBody)
	if d, ok := body["done"].(bool); !ok || d {
		t.Errorf("in-progress status should send done=false; got %v", body["done"])
	}
	for _, k := range []string{"description", "priority", "title"} {
		if _, present := body[k]; present {
			t.Errorf("status-only update must not send %q (#2962: it would clobber the stored value); body=%s", k, *patchBody)
		}
	}
}

func decodePatchBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	if len(raw) == 0 {
		t.Fatal("no PATCH body was captured")
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode patch body %q: %v", string(raw), err)
	}
	return body
}
