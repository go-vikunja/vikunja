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
