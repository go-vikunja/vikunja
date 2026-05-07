package client

import (
	"context"
	"fmt"
)

// CreateRelation links two tasks. relationKind is "subtask", "parenttask",
// "blocking", "blocked", "related", etc.
func (c *Client) CreateRelation(ctx context.Context, taskID int64, otherTaskID int64, relationKind string) (*TaskRelation, error) {
	var out TaskRelation
	body := &TaskRelation{OtherTaskID: otherTaskID, RelationKind: relationKind}
	if err := c.Do(ctx, "PUT", fmt.Sprintf("/tasks/%d/relations", taskID), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
