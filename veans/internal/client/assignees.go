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

// AddAssignee assigns a user (typically the bot) to a task.
func (c *Client) AddAssignee(ctx context.Context, taskID, userID int64) error {
	return c.Do(ctx, "PUT", fmt.Sprintf("/tasks/%d/assignees", taskID), nil, &TaskAssignee{UserID: userID}, nil)
}

// RemoveAssignee unassigns.
func (c *Client) RemoveAssignee(ctx context.Context, taskID, userID int64) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/tasks/%d/assignees/%d", taskID, userID), nil, nil, nil)
}
