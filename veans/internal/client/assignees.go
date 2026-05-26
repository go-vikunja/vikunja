package client

import (
	"context"
	"fmt"
)

// AddAssignee assigns a user (typically the bot) to a task.
func (c *Client) AddAssignee(ctx context.Context, taskID, userID int64) error {
	return c.Do(ctx, "PUT", fmt.Sprintf("/tasks/%d/assignees", taskID), nil, &TaskAssignee{UserID: userID}, nil)
}

// RemoveAssignee unassigns.
func (c *Client) RemoveAssignee(ctx context.Context, taskID, userID int64) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/tasks/%d/assignees/%d", taskID, userID), nil, nil, nil)
}
