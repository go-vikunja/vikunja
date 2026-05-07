package client

import (
	"context"
	"fmt"
)

// ListBuckets returns the buckets configured on a Kanban view.
func (c *Client) ListBuckets(ctx context.Context, projectID, viewID int64) ([]*Bucket, error) {
	var out []*Bucket
	path := fmt.Sprintf("/projects/%d/views/%d/buckets", projectID, viewID)
	if err := c.Do(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateBucket inserts a new bucket into a Kanban view.
func (c *Client) CreateBucket(ctx context.Context, projectID, viewID int64, b *Bucket) (*Bucket, error) {
	var out Bucket
	path := fmt.Sprintf("/projects/%d/views/%d/buckets", projectID, viewID)
	if b == nil {
		b = &Bucket{}
	}
	b.ProjectViewID = viewID
	if err := c.Do(ctx, "PUT", path, nil, b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
