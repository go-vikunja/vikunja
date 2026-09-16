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
)

// ListBuckets returns the buckets configured on a Kanban view. Bucket.ReadAll
// ignores page/per_page and returns every bucket in a single page (the envelope
// total reflects the full set), so one GET gets them all — paging would
// re-fetch the same buckets and duplicate them. Unwrap .items.
func (c *Client) ListBuckets(ctx context.Context, projectID, viewID int64) ([]*Bucket, error) {
	path := fmt.Sprintf("/projects/%d/views/%d/buckets", projectID, viewID)
	items, _, err := doList[*Bucket](ctx, c, path, nil)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// CreateBucket inserts a new bucket into a Kanban view. The project and view
// come from the URL; the v2 handler ignores project_view_id in the body.
func (c *Client) CreateBucket(ctx context.Context, projectID, viewID int64, b *Bucket) (*Bucket, error) {
	var out Bucket
	path := fmt.Sprintf("/projects/%d/views/%d/buckets", projectID, viewID)
	if b == nil {
		b = &Bucket{}
	}
	if err := c.Do(ctx, "POST", path, nil, b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MoveTaskToBucket positions an existing task in `bucketID` on the
// project's view. Vikunja stores task↔bucket relations in a separate
// table (`task_buckets`); a task update with bucket_id does not reliably
// move tasks — this dedicated endpoint is the one the Kanban UI's
// drag-and-drop uses. On v2 it's a PUT, and project/view/bucket all come
// from the URL, so the body only carries the task id.
func (c *Client) MoveTaskToBucket(ctx context.Context, projectID, viewID, bucketID, taskID int64) error {
	path := fmt.Sprintf("/projects/%d/views/%d/buckets/%d/tasks",
		projectID, viewID, bucketID)
	body := map[string]int64{"task_id": taskID}
	return c.Do(ctx, "PUT", path, nil, body, nil)
}
