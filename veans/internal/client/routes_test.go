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
	"slices"
	"testing"
)

func TestPermissionsForBot_DropsUnknownGroups(t *testing.T) {
	// Server only exposes a subset of what we ask for.
	server := map[string]RouteGroup{
		"tasks": {
			"read_one": {},
			"read_all": {},
			"create":   {},
			"update":   {},
			// "delete" intentionally absent
		},
		"projects": {
			"read_one": {},
			"read_all": {},
		},
		// no "labels", no "comments", etc.
	}
	got := PermissionsForBot(server)

	if _, ok := got["tasks"]; !ok {
		t.Fatalf("expected tasks group in result")
	}
	for _, a := range got["tasks"] {
		if a == "delete" {
			t.Errorf("delete should have been dropped")
		}
	}
	if _, ok := got["projects"]; !ok {
		t.Fatalf("expected projects group")
	}
	if _, ok := got["labels"]; ok {
		t.Errorf("labels was not on server, should not appear in result")
	}
	if _, ok := got["nonexistent_group"]; ok {
		t.Errorf("phantom group leaked into result")
	}
}

func TestPermissionsForBot_EmptyWhenServerIsEmpty(t *testing.T) {
	got := PermissionsForBot(map[string]RouteGroup{})
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}

// Pins current (create/update/tasks_update) and legacy (post/put/tasks_put) bucket keys.
func TestPermissionsForBot_ProjectsBucketScopes(t *testing.T) {
	server := map[string]RouteGroup{
		"projects": {
			"read_one":                   {},
			"read_all":                   {},
			"tasks_by-index":             {},
			"views_buckets":              {}, // list buckets
			"views_buckets_post":         {}, // create bucket
			"views_buckets_put":          {}, // update bucket
			"views_buckets_delete":       {}, // delete bucket
			"views_buckets_tasks":        {}, // buckets-with-tasks list
			"views_buckets_tasks_put":    {}, // legacy move
			"views_buckets_create":       {},
			"views_buckets_update":       {},
			"views_buckets_tasks_update": {}, // move
		},
	}
	got := PermissionsForBot(server)
	projects, ok := got["projects"]
	if !ok {
		t.Fatalf("expected projects group in result")
	}
	want := []string{
		"read_one", "read_all", "tasks_by-index",
		"views_buckets", "views_buckets_post", "views_buckets_put",
		"views_buckets_delete", "views_buckets_tasks", "views_buckets_tasks_put",
		"views_buckets_create", "views_buckets_update", "views_buckets_tasks_update",
	}
	for _, w := range want {
		if !slices.Contains(projects, w) {
			t.Errorf("projects scope %q missing from bot grant; got %v", w, projects)
		}
	}
}
