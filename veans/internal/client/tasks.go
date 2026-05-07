package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// TaskListOptions selects which tasks to return from ListProjectTasks.
type TaskListOptions struct {
	Filter     string
	Page       int
	PerPage    int
	SortBy     []string
	OrderBy    []string
	Expand     []string
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

// GetTask fetches a single task by numeric ID.
func (c *Client) GetTask(ctx context.Context, id int64) (*Task, error) {
	var out Task
	if err := c.Do(ctx, "GET", fmt.Sprintf("/tasks/%d", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
