package client

import (
	"context"
	"fmt"
)

// AddTaskComment posts a new comment on a task.
func (c *Client) AddTaskComment(ctx context.Context, taskID int64, body string) (*TaskComment, error) {
	var out TaskComment
	if err := c.Do(ctx, "PUT", fmt.Sprintf("/tasks/%d/comments", taskID), nil, &TaskComment{Comment: body}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTaskComments returns all comments on a task.
func (c *Client) ListTaskComments(ctx context.Context, taskID int64) ([]*TaskComment, error) {
	var out []*TaskComment
	if err := c.Do(ctx, "GET", fmt.Sprintf("/tasks/%d/comments", taskID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
