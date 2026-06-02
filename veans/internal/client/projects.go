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

// ListProjects pages through GET /projects, accumulating until the server's
// x-pagination-total-pages header says we're done.
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var all []*Project
	page := 1
	for {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "50")
		var batch []*Project
		total, err := c.DoPaginated(ctx, "GET", "/projects", q, &batch)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if paginationDone(page, len(batch), 50, total) {
			return all, nil
		}
		page++
	}
}

// GetProject fetches a single project by ID.
func (c *Client) GetProject(ctx context.Context, id int64) (*Project, error) {
	var out Project
	if err := c.Do(ctx, "GET", fmt.Sprintf("/projects/%d", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProject creates a new project owned by the calling user. Vikunja
// auto-creates the default views (List, Gantt, Table, Kanban) on insert.
func (c *Client) CreateProject(ctx context.Context, p *Project) (*Project, error) {
	var out Project
	if err := c.Do(ctx, "PUT", "/projects", nil, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ShareProjectWithUser grants `username` `permission` on project `id`.
func (c *Client) ShareProjectWithUser(ctx context.Context, projectID int64, share *ProjectUser) (*ProjectUser, error) {
	var out ProjectUser
	if err := c.Do(ctx, "PUT", fmt.Sprintf("/projects/%d/users", projectID), nil, share, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListProjectViews returns saved views (Kanban, List, …) on a project.
func (c *Client) ListProjectViews(ctx context.Context, projectID int64) ([]*ProjectView, error) {
	var out []*ProjectView
	if err := c.Do(ctx, "GET", fmt.Sprintf("/projects/%d/views", projectID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
