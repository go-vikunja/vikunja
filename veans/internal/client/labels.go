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
