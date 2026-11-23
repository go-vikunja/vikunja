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
	"strconv"
	"testing"

	"code.vikunja.io/veans/internal/client"
)

func decodeTask(t *testing.T, out string) *client.Task {
	t.Helper()
	var task client.Task
	if err := json.Unmarshal([]byte(out), &task); err != nil {
		t.Fatalf("decode task: %v\n%s", err, out)
	}
	return &task
}

func hasParent(task *client.Task, parentID int64) bool {
	for _, rt := range task.RelatedTasks["parenttask"] {
		if rt != nil && rt.ID == parentID {
			return true
		}
	}
	return false
}

func TestCreate_NonTodoStatus(t *testing.T) {
	ws, h := provisionWorkspace(t)
	cfg := loadConfig(t, ws)

	parentOut, errOut, code := h.Run(t, ws, "create", "parent")
	if code != 0 {
		t.Fatalf("create parent exit %d\n%s", code, errOut)
	}
	parent := decodeTask(t, parentOut)

	for status, want := range map[string]int64{
		"in-progress": cfg.Buckets.InProgress,
		"in-review":   cfg.Buckets.InReview,
	} {
		t.Run(status, func(t *testing.T) {
			out, errOut, code := h.Run(t, ws, "create", "probe "+status,
				"-s", status, "-d", "<p>x</p>", "--parent", strconv.FormatInt(parent.Index, 10))
			if code != 0 {
				t.Fatalf("create -s %s exit %d\nstdout: %s\nstderr: %s", status, code, out, errOut)
			}
			created := decodeTask(t, out)
			server := h.GetTask(t, created.ID)
			if got := server.CurrentBucketID(cfg.ViewID); got != want {
				t.Fatalf("bucket = %d, want %d", got, want)
			}
			if !hasParent(server, parent.ID) {
				t.Fatalf("parent relation missing: %+v", server.RelatedTasks)
			}
		})
	}
}

// Vikunja un-marks a task leaving the done bucket, so scrapped lands in the
// Scrapped bucket with done=false, even when the task was completed before.
func TestUpdate_Scrapped(t *testing.T) {
	ws, h := provisionWorkspace(t)
	cfg := loadConfig(t, ws)

	for _, from := range []string{"todo", "completed"} {
		t.Run("from "+from, func(t *testing.T) {
			out, errOut, code := h.Run(t, ws, "create", "to scrap", "-s", from)
			if code != 0 {
				t.Fatalf("create exit %d\n%s", code, errOut)
			}
			created := decodeTask(t, out)

			out, errOut, code = h.Run(t, ws, "update", strconv.FormatInt(created.Index, 10),
				"-s", "scrapped", "--reason", "probe")
			if code != 0 {
				t.Fatalf("update exit %d\nstdout: %s\nstderr: %s", code, out, errOut)
			}
			printed := decodeTask(t, out)
			server := h.GetTask(t, created.ID)
			if got := server.CurrentBucketID(cfg.ViewID); got != cfg.Buckets.Scrapped {
				t.Fatalf("bucket = %d, want scrapped %d", got, cfg.Buckets.Scrapped)
			}
			if server.Done || printed.Done {
				t.Fatalf("scrapped task should not be done (server=%v printed=%v)", server.Done, printed.Done)
			}
		})
	}
}

func TestListReady_ScrappedBlockerDoesNotBlock(t *testing.T) {
	ws, h := provisionWorkspace(t)

	out, errOut, code := h.Run(t, ws, "create", "blocker")
	if code != 0 {
		t.Fatalf("create blocker exit %d\n%s", code, errOut)
	}
	blocker := decodeTask(t, out)
	out, errOut, code = h.Run(t, ws, "create", "dependent", "--blocked-by", strconv.FormatInt(blocker.Index, 10))
	if code != 0 {
		t.Fatalf("create dependent exit %d\n%s", code, errOut)
	}
	dependent := decodeTask(t, out)

	if _, errOut, code = h.Run(t, ws, "update", strconv.FormatInt(blocker.Index, 10),
		"-s", "scrapped", "--reason", "probe"); code != 0 {
		t.Fatalf("scrap blocker exit %d\n%s", code, errOut)
	}

	out, errOut, code = h.Run(t, ws, "list", "--ready")
	if code != 0 {
		t.Fatalf("list --ready exit %d\n%s", code, errOut)
	}
	var ready []*client.Task
	if err := json.Unmarshal([]byte(out), &ready); err != nil {
		t.Fatal(err)
	}
	for _, r := range ready {
		if r.ID == dependent.ID {
			return
		}
	}
	t.Fatalf("dependent of a scrapped blocker should be ready; got %d ready tasks", len(ready))
}

func TestUpdate_Parent(t *testing.T) {
	ws, h := provisionWorkspace(t)

	mk := func(title string) *client.Task {
		out, errOut, code := h.Run(t, ws, "create", title)
		if code != 0 {
			t.Fatalf("create %q exit %d\n%s", title, code, errOut)
		}
		return decodeTask(t, out)
	}
	first, second, child := mk("first parent"), mk("second parent"), mk("child")
	childRef := strconv.FormatInt(child.Index, 10)

	update := func(args ...string) *client.Task {
		t.Helper()
		out, errOut, code := h.Run(t, ws, append([]string{"update", childRef}, args...)...)
		if code != 0 {
			t.Fatalf("update %v exit %d\nstdout: %s\nstderr: %s", args, code, out, errOut)
		}
		return decodeTask(t, out)
	}

	printed := update("--parent", strconv.FormatInt(first.Index, 10))
	if !hasParent(printed, first.ID) {
		t.Fatalf("parent not set: %+v", printed.RelatedTasks)
	}

	// Setting a new parent replaces the old one.
	update("--parent", strconv.FormatInt(second.Index, 10))
	server := h.GetTask(t, child.ID)
	if !hasParent(server, second.ID) || hasParent(server, first.ID) {
		t.Fatalf("parent not replaced: %+v", server.RelatedTasks["parenttask"])
	}

	update("--no-parent")
	server = h.GetTask(t, child.ID)
	if n := len(server.RelatedTasks["parenttask"]); n != 0 {
		t.Fatalf("parent not removed: %+v", server.RelatedTasks["parenttask"])
	}
}

func TestUpdate_BlockedBy(t *testing.T) {
	ws, h := provisionWorkspace(t)

	mk := func(title string) *client.Task {
		out, errOut, code := h.Run(t, ws, "create", title)
		if code != 0 {
			t.Fatalf("create %q exit %d\n%s", title, code, errOut)
		}
		return decodeTask(t, out)
	}
	a, b, task := mk("blocker a"), mk("blocker b"), mk("blocked")
	ref := func(x *client.Task) string { return strconv.FormatInt(x.Index, 10) }
	blockers := func() map[int64]bool {
		got := map[int64]bool{}
		for _, r := range h.GetTask(t, task.ID).RelatedTasks["blocked"] {
			got[r.ID] = true
		}
		return got
	}
	update := func(args ...string) {
		t.Helper()
		if _, errOut, code := h.Run(t, ws, append([]string{"update", ref(task)}, args...)...); code != 0 {
			t.Fatalf("update %v exit %d\n%s", args, code, errOut)
		}
	}

	update("--blocked-by-add", ref(a), "--blocked-by-add", ref(b))
	if got := blockers(); !got[a.ID] || !got[b.ID] || len(got) != 2 {
		t.Fatalf("after add: %v", got)
	}

	// Re-adding is a no-op; removing a non-blocker too.
	update("--blocked-by-add", ref(a), "--blocked-by-remove", ref(b), "--blocked-by-remove", ref(task))
	if got := blockers(); !got[a.ID] || len(got) != 1 {
		t.Fatalf("after remove: %v", got)
	}
}
