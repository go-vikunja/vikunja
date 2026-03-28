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

package mcp

import "time"

type TaskFilter struct {
	ProjectID *int64  `json:"project_id,omitempty"`
	ListID    *int64  `json:"list_id,omitempty"`
	IsDone    *bool   `json:"is_done,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	Offset    *int    `json:"offset,omitempty"`
	Search    *string `json:"search,omitempty"`
}

type CreateTaskInput struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	ProjectID   int64      `json:"project_id"`
	ListID      *int64     `json:"list_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    *int64     `json:"priority,omitempty"`
	Labels      []int64    `json:"labels,omitempty"`
	Assignees   []int64    `json:"assignees,omitempty"`
}

type UpdateTaskInput struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Done        *bool      `json:"done,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    *int64     `json:"priority,omitempty"`
	Labels      []int64    `json:"labels,omitempty"`
	Assignees   []int64    `json:"assignees,omitempty"`
}

type ProjectFilter struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

type ListFilter struct {
	ProjectID int64 `json:"project_id"`
	Limit     *int  `json:"limit,omitempty"`
	Offset    *int  `json:"offset,omitempty"`
}

type KanbanBoardFilter struct {
	ProjectID int64 `json:"project_id"`
}
