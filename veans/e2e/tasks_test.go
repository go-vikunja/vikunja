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

package e2e

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/veans/internal/client"
)

// TestCreateShowList_RoundTrip verifies the read+write path against a real
// Vikunja: provision a workspace via init, create a task, show it, list it
// (with --filter), and confirm the JSON shapes are unwrapped raw object/array.
func TestCreateShowList_RoundTrip(t *testing.T) {
	ws, h := provisionWorkspace(t)

	// Create a task with a description and a label.
	out, errOut, code := h.Run(t, ws,
		"create", "Test bug fix",
		"-d", "## Repro\n- [ ] step 1\n- [ ] step 2",
		"--label", "bug",
		"--priority", "3",
	)
	if code != 0 {
		t.Fatalf("create exit %d\n%s\n%s", code, out, errOut)
	}
	var created client.Task
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("decode create: %v\n%s", err, out)
	}
	if created.Title != "Test bug fix" {
		t.Fatalf("created title = %q", created.Title)
	}
	if created.Priority != 3 {
		t.Fatalf("priority = %d", created.Priority)
	}
	if !taskHasLabelTitle(&created, "veans:bug") {
		t.Fatalf("expected veans:bug label on created task; got %+v", created.Labels)
	}

	// Show with --json — should be a raw object, not enveloped.
	id := fmt.Sprintf("%d", created.Index)
	showOut, _, code := h.Run(t, ws, "show", id)
	if code != 0 {
		t.Fatalf("show exit %d", code)
	}
	var shown client.Task
	if err := json.Unmarshal([]byte(showOut), &shown); err != nil {
		t.Fatalf("decode show: %v\n%s", err, showOut)
	}
	if shown.ID != created.ID {
		t.Fatalf("show returned wrong task: %d vs %d", shown.ID, created.ID)
	}

	// List with --json — should be a raw array.
	listOut, _, code := h.Run(t, ws, "list")
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
	var listed []*client.Task
	if err := json.Unmarshal([]byte(listOut), &listed); err != nil {
		t.Fatalf("decode list: %v\n%s", err, listOut)
	}
	if len(listed) == 0 {
		t.Fatalf("list returned empty array; expected at least our created task")
	}

	// --filter passthrough: only items with priority > 2.
	filterOut, _, code := h.Run(t, ws, "list", "--filter", "priority > 2")
	if code != 0 {
		t.Fatalf("list --filter exit %d\n%s", code, filterOut)
	}
	var filtered []*client.Task
	if err := json.Unmarshal([]byte(filterOut), &filtered); err != nil {
		t.Fatalf("decode filter list: %v", err)
	}
	for _, ft := range filtered {
		if ft.Priority <= 2 {
			t.Fatalf("filter leaked priority=%d task into result", ft.Priority)
		}
	}

	// Empty result set must encode as `[]`, not `null` — JSON-parsing agents
	// can't reliably branch on the latter. Use a filter that matches nothing.
	emptyOut, _, code := h.Run(t, ws, "list", "--filter", "priority > 10")
	if code != 0 {
		t.Fatalf("list (empty) exit %d\n%s", code, emptyOut)
	}
	if got := strings.TrimSpace(emptyOut); got != "[]" {
		t.Fatalf("empty list should print []; got %q", got)
	}
}

// TestUpdate_DescriptionReplaceUniqueness pins the agent-friendly Edit-tool
// behavior: the "old" string must match exactly once, otherwise the update
// errors and nothing changes.
func TestUpdate_DescriptionReplaceUniqueness(t *testing.T) {
	ws, h := provisionWorkspace(t)

	out, errOut, code := h.Run(t, ws, "create", "checkbox task",
		"-d", "- [ ] step 1\n- [ ] step 1 (again)",
	)
	if code != 0 {
		t.Fatalf("create exit %d\n%s\n%s", code, out, errOut)
	}
	var created client.Task
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatal(err)
	}
	id := fmt.Sprintf("%d", created.Index)

	// Non-unique replace should fail with a validation error and a
	// non-zero exit code.
	_, stderr, code := h.Run(t, ws,
		"update", id,
		"--description-replace-old", "step 1",
		"--description-replace-new", "step 1 [x]",
	)
	if code == 0 {
		t.Fatalf("expected non-zero exit on non-unique replace")
	}
	if !strings.Contains(stderr, "VALIDATION_ERROR") {
		t.Fatalf("expected VALIDATION_ERROR in stderr, got: %s", stderr)
	}

	// Disambiguate and try again — should succeed.
	_, _, code = h.Run(t, ws,
		"update", id,
		"--description-replace-old", "- [ ] step 1\n",
		"--description-replace-new", "- [x] step 1\n",
	)
	if code != 0 {
		t.Fatalf("disambiguated update should succeed; exit %d", code)
	}
}

func taskHasLabelTitle(t *client.Task, title string) bool {
	for _, l := range t.Labels {
		if l != nil && l.Title == title {
			return true
		}
	}
	return false
}

// TestFields_AllCommands pins --fields on every task-printing command and
// the default stripping of nested related-task descriptions.
func TestFields_AllCommands(t *testing.T) {
	ws, h := provisionWorkspace(t)

	out, errOut, code := h.Run(t, ws, "create", "parent", "-d", "<p>parent body</p>")
	if code != 0 {
		t.Fatalf("create parent exit %d\n%s", code, errOut)
	}
	parent := decodeTask(t, out)
	parentRef := strconv.FormatInt(parent.Index, 10)

	out, errOut, code = h.Run(t, ws, "create", "child", "--parent", parentRef, "--fields", "identifier,url")
	if code != 0 {
		t.Fatalf("create --fields exit %d\n%s", code, errOut)
	}
	var created map[string]any
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("decode %s: %v", out, err)
	}
	if len(created) != 2 || created["identifier"] == nil || created["url"] == nil {
		t.Fatalf("create --fields should print only identifier and url: %s", out)
	}
	childRef := strings.TrimPrefix(created["identifier"].(string), loadConfig(t, ws).ProjectIdentifier+"-")

	for _, args := range [][]string{
		{"show", childRef},
		{"update", childRef, "--priority", "2"},
		{"claim", childRef},
	} {
		out, errOut, code := h.Run(t, ws, append(args, "--fields", "identifier,url")...)
		if code != 0 {
			t.Fatalf("%v exit %d\n%s", args, code, errOut)
		}
		if got := strings.TrimSpace(out); !strings.HasPrefix(got, `{"identifier":`) || strings.Count(got, `"url"`) != 1 {
			t.Fatalf("%v --fields: %s", args, got)
		}
	}

	out, errOut, code = h.Run(t, ws, "list", "--fields", "identifier")
	if code != 0 {
		t.Fatalf("list --fields exit %d\n%s", code, errOut)
	}
	var listed []map[string]any
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || len(listed[0]) != 1 {
		t.Fatalf("list --fields: %s", out)
	}

	if _, errOut, code = h.Run(t, ws, "show", childRef, "--fields", "bogus"); code == 0 || !strings.Contains(errOut, "VALIDATION_ERROR") {
		t.Fatalf("unknown field should fail validation: exit %d %s", code, errOut)
	}

	// Default output: whole task, nested parent without its description.
	out, _, _ = h.Run(t, ws, "show", childRef)
	if strings.Contains(out, "parent body") {
		t.Fatalf("nested related-task description leaked: %s", out)
	}
	if !hasParent(decodeTask(t, out), parent.ID) {
		t.Fatalf("show lost the relation: %s", out)
	}
}

// TestTaskIDFlag pins --id on show/update/claim and the not-found hint
// when a numeric task id is passed where an index is expected.
func TestTaskIDFlag(t *testing.T) {
	ws, h := provisionWorkspace(t)

	out, errOut, code := h.Run(t, ws, "create", "by id")
	if code != 0 {
		t.Fatalf("create exit %d\n%s", code, errOut)
	}
	created := decodeTask(t, out)
	id := strconv.FormatInt(created.ID, 10)

	for _, args := range [][]string{
		{"show", "--id", id},
		{"update", "--id", id, "--priority", "3"},
		{"claim", "--id", id},
	} {
		out, errOut, code := h.Run(t, ws, append(args, "--fields", "id")...)
		if code != 0 {
			t.Fatalf("%v exit %d\n%s", args, code, errOut)
		}
		if got := strings.TrimSpace(out); got != `{"id":`+id+`}` {
			t.Fatalf("%v printed %s", args, got)
		}
	}

	_, errOut, code = h.Run(t, ws, "show", "999999")
	if code == 0 || !strings.Contains(errOut, "NOT_FOUND") || !strings.Contains(errOut, "--id 999999") {
		t.Fatalf("bare-number not-found should point at --id: exit %d %s", code, errOut)
	}
	_, errOut, code = h.Run(t, ws, "show", "1", "--id", id)
	if code == 0 || !strings.Contains(errOut, "VALIDATION_ERROR") {
		t.Fatalf("ref plus --id should fail validation: exit %d %s", code, errOut)
	}
}
