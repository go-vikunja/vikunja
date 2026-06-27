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
	"fmt"
	"net/url"
	"strconv"
)

// TaskListOptions selects which tasks to return from ListProjectTasks.
type TaskListOptions struct {
	Filter  string
	Page    int
	PerPage int
	Expand  []string
}

func (o *TaskListOptions) values() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	if o.Filter != "" {
		q.Set("filter", o.Filter)
	}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(o.PerPage))
	}
	for _, e := range o.Expand {
		q.Add("expand", e)
	}
	return q
}

// ListProjectTasks paginates `GET /projects/{id}/tasks` exhaustively,
// terminating against the list envelope's total_pages.
func (c *Client) ListProjectTasks(ctx context.Context, projectID int64, opts *TaskListOptions) ([]*Task, error) {
	if opts == nil {
		opts = &TaskListOptions{}
	}
	per := opts.PerPage
	if per <= 0 {
		per = 50
	}
	path := fmt.Sprintf("/projects/%d/tasks", projectID)
	var all []*Task
	page := 1
	for {
		o := *opts
		o.Page = page
		o.PerPage = per
		batch, totalPages, err := doList[*Task](ctx, c, path, o.values())
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if page >= totalPages {
			return all, nil
		}
		page++
	}
}

// GetTask fetches a single task by numeric ID. expand=buckets is requested
// because Vikunja's bare GET returns bucket_id=0 — the per-view bucket
// memberships only surface under the Buckets slice.
func (c *Client) GetTask(ctx context.Context, id int64) (*Task, error) {
	var out Task
	q := url.Values{}
	q.Add("expand", "buckets")
	if err := c.Do(ctx, "GET", fmt.Sprintf("/tasks/%d", id), q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrentBucketID returns the task's bucket id on the given project view,
// or 0 if no bucket entry is present (which happens when buckets aren't
// expanded, or the task is in no view-bound bucket yet).
func (t *Task) CurrentBucketID(viewID int64) int64 {
	if t.BucketID != 0 {
		return t.BucketID
	}
	for _, b := range t.Buckets {
		if b == nil {
			continue
		}
		// Buckets returned via expand=buckets are scoped to the requesting
		// view; without view scoping the slice can include entries from
		// every view this task belongs to.
		if viewID == 0 || b.ProjectViewID == viewID || b.ProjectViewID == 0 {
			return b.ID
		}
	}
	return 0
}

// CreateTask inserts a task into a project (POST /projects/{id}/tasks).
func (c *Client) CreateTask(ctx context.Context, projectID int64, t *Task) (*Task, error) {
	var out Task
	if err := c.Do(ctx, "POST", fmt.Sprintf("/projects/%d/tasks", projectID), nil, t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTask partially updates a task via PATCH /tasks/{id} with a JSON Merge
// Patch body: only the fields set on `patch` are written, the rest are left
// intact (the fix for issue #2962, where a status-only update used to zero
// description and priority). This endpoint does NOT move tasks between
// buckets — the task↔bucket relation is row-shaped in task_buckets, and
// bucket_id on the request body is ignored. Use MoveTaskToBucket() for that.
// The server still auto-flips the bucket when `done` toggles, between the
// canonical "todo" and "done" buckets the project view is configured with.
func (c *Client) UpdateTask(ctx context.Context, id int64, patch *TaskPatch) (*Task, error) {
	var out Task
	if err := c.DoMerge(ctx, "PATCH", fmt.Sprintf("/tasks/%d", id), nil, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
