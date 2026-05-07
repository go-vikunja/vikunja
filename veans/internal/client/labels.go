package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListLabels paginates GET /labels and returns every label visible to the
// authenticated user (labels are global per user, not scoped to a project).
func (c *Client) ListLabels(ctx context.Context, search string) ([]*Label, error) {
	var all []*Label
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "50")
		if search != "" {
			q.Set("s", search)
		}
		var batch []*Label
		if err := c.Do(ctx, "GET", "/labels", q, nil, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 50 {
			return all, nil
		}
	}
}

// CreateLabel creates a new label owned by the authenticated user.
func (c *Client) CreateLabel(ctx context.Context, l *Label) (*Label, error) {
	var out Label
	if err := c.Do(ctx, "PUT", "/labels", nil, l, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddLabelToTask attaches an existing label to a task.
func (c *Client) AddLabelToTask(ctx context.Context, taskID, labelID int64) error {
	return c.Do(ctx, "PUT", fmt.Sprintf("/tasks/%d/labels", taskID), nil, &LabelTask{LabelID: labelID}, nil)
}

// RemoveLabelFromTask detaches a label.
func (c *Client) RemoveLabelFromTask(ctx context.Context, taskID, labelID int64) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/tasks/%d/labels/%d", taskID, labelID), nil, nil, nil)
}
