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
	"testing"

	"code.vikunja.io/veans/internal/client"
)

func TestIsReady(t *testing.T) {
	const todoBucket int64 = 11
	const viewID int64 = 7

	cases := []struct {
		name string
		task *client.Task
		want bool
	}{
		{
			name: "in todo, no relations -> ready",
			task: &client.Task{
				ID:      1,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
			},
			want: true,
		},
		{
			name: "subtask with only parenttask + blocking -> ready",
			task: &client.Task{
				ID:      2,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
				RelatedTasks: map[string][]*client.Task{
					"parenttask": {{ID: 1, Done: false}},
					"blocking":   {{ID: 3, Done: false}},
				},
			},
			want: true,
		},
		{
			name: "blocked by incomplete task -> not ready",
			task: &client.Task{
				ID:      3,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
				RelatedTasks: map[string][]*client.Task{
					"blocked": {{ID: 2, Done: false}},
				},
			},
			want: false,
		},
		{
			name: "blocked by completed task -> ready",
			task: &client.Task{
				ID:      3,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
				RelatedTasks: map[string][]*client.Task{
					"blocked": {{ID: 2, Done: true}},
				},
			},
			want: true,
		},
		{
			name: "done task -> not ready",
			task: &client.Task{
				ID:      4,
				Done:    true,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
			},
			want: false,
		},
		{
			name: "in another bucket -> not ready",
			task: &client.Task{
				ID:      5,
				Buckets: []*client.Bucket{{ID: 99, ProjectViewID: viewID}},
			},
			want: false,
		},
		{
			name: "blocked by mix of done and incomplete -> not ready",
			task: &client.Task{
				ID:      6,
				Buckets: []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}},
				RelatedTasks: map[string][]*client.Task{
					"blocked": {{ID: 7, Done: true}, {ID: 8, Done: false}},
				},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isReady(tc.task, todoBucket, viewID); got != tc.want {
				t.Fatalf("isReady = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestIsReady_SubtaskWithBlockedSibling pins the scenario from the bug
// report: two tasks share a parent; one has no "blocked" entry and is
// ready; the other is blocked by the first and is not ready. Once the
// first completes, the second becomes ready.
func TestIsReady_SubtaskWithBlockedSibling(t *testing.T) {
	const todoBucket int64 = 11
	const viewID int64 = 7
	bucket := []*client.Bucket{{ID: todoBucket, ProjectViewID: viewID}}

	// Task #2: subtask of #1, blocks #3. No "blocked" relations of its own.
	taskTwo := &client.Task{
		ID: 2, Done: false, Buckets: bucket,
		RelatedTasks: map[string][]*client.Task{
			"parenttask": {{ID: 1, Done: false}},
			"blocking":   {{ID: 3, Done: false}},
		},
	}
	// Task #3: subtask of #1, blocked by #2.
	taskThree := &client.Task{
		ID: 3, Done: false, Buckets: bucket,
		RelatedTasks: map[string][]*client.Task{
			"parenttask": {{ID: 1, Done: false}},
			"blocked":    {{ID: 2, Done: false}},
		},
	}

	if !isReady(taskTwo, todoBucket, viewID) {
		t.Fatal("task #2 should be ready (no incomplete blockers)")
	}
	if isReady(taskThree, todoBucket, viewID) {
		t.Fatal("task #3 should not be ready (blocked by incomplete #2)")
	}

	// Now complete #2 and reflect that in #3's relation snapshot.
	taskThree.RelatedTasks["blocked"][0].Done = true
	if !isReady(taskThree, todoBucket, viewID) {
		t.Fatal("task #3 should be ready once #2 is done")
	}
}
