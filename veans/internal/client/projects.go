package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListProjects pages through GET /projects, accumulating until exhausted.
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var all []*Project
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "50")
		var batch []*Project
		if err := c.Do(ctx, "GET", "/projects", q, nil, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 50 {
			return all, nil
		}
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
