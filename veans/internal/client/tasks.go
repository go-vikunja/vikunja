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
	SortBy  []string
	OrderBy []string
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
	for _, s := range o.SortBy {
		q.Add("sort_by", s)
	}
	for _, s := range o.OrderBy {
		q.Add("order_by", s)
	}
	for _, e := range o.Expand {
		q.Add("expand", e)
	}
	return q
}

// ListProjectTasks paginates `GET /projects/{id}/tasks` exhaustively.
func (c *Client) ListProjectTasks(ctx context.Context, projectID int64, opts *TaskListOptions) ([]*Task, error) {
	if opts == nil {
		opts = &TaskListOptions{}
	}
	per := opts.PerPage
	if per <= 0 {
		per = 50
	}
	var all []*Task
	for page := 1; ; page++ {
		o := *opts
		o.Page = page
		o.PerPage = per
		var batch []*Task
		if err := c.Do(ctx, "GET", fmt.Sprintf("/projects/%d/tasks", projectID), o.values(), nil, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < per {
			return all, nil
		}
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

// CreateTask inserts a task into a project (PUT /projects/{id}/tasks).
func (c *Client) CreateTask(ctx context.Context, projectID int64, t *Task) (*Task, error) {
	var out Task
	if err := c.Do(ctx, "PUT", fmt.Sprintf("/projects/%d/tasks", projectID), nil, t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTask updates a task (POST /tasks/{id}). bucket_id moves the task
// between buckets in the same view.
func (c *Client) UpdateTask(ctx context.Context, id int64, t *Task) (*Task, error) {
	var out Task
	if err := c.Do(ctx, "POST", fmt.Sprintf("/tasks/%d", id), nil, t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
