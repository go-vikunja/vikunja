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

import "context"

// RouteGroup mirrors models.APITokenRoute on the wire — the per-action
// detail object is opaque to us.
type RouteGroup map[string]struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// Routes returns the API token route map. Used during bootstrap to
// negotiate exactly which permission groups+actions exist on this Vikunja
// instance, so the bot's API token only requests scopes the server knows
// about — avoiding hard-coding a permission list that could drift.
func (c *Client) Routes(ctx context.Context) (map[string]RouteGroup, error) {
	out := map[string]RouteGroup{}
	if err := c.Do(ctx, "GET", "/routes", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PermissionsForBot picks a curated subset of route groups the veans bot
// needs and projects the available actions of each. Groups not present on
// the server are silently dropped, so the resulting permission map is
// always valid for PUT /tokens regardless of Vikunja version.
//
// The action names reflect Vikunja's actual route map (see GET /routes):
// bucket CRUD and the bucket-task move endpoint live under the `projects`
// group as `views_buckets*` and `views_buckets_tasks`, not a separate
// `buckets` group.
func PermissionsForBot(routes map[string]RouteGroup) map[string][]string {
	wanted := map[string][]string{
		// Read + write tasks across the project. The bot creates, updates,
		// and reads tasks; it doesn't delete (humans/merge hook close).
		"tasks": {
			"read_one", "read_all", "create", "update", "position",
			"read", "update_bulk",
		},
		// Project access: read project metadata, manage buckets & move
		// tasks between them. tasks_by-index resolves #NN / PROJ-NN.
		"projects": {
			"read_one", "read_all", "tasks_by-index",
			"views_buckets", "views_buckets_put", "views_buckets_post",
			"views_buckets_delete", "views_buckets_tasks",
		},
		"projects_views":  {"read_one", "read_all"},
		"labels":          {"read_one", "read_all", "create", "update", "delete"},
		"tasks_comments":  {"read_one", "read_all", "create", "update", "delete"},
		"tasks_relations": {"create", "delete"},
		"tasks_assignees": {"read_all", "create", "delete", "update_bulk"},
		"tasks_labels":    {"create", "delete", "read_all", "update_bulk"},
	}
	out := map[string][]string{}
	for group, actions := range wanted {
		avail, ok := routes[group]
		if !ok {
			continue
		}
		var picked []string
		for _, a := range actions {
			if _, has := avail[a]; has {
				picked = append(picked, a)
			}
		}
		if len(picked) > 0 {
			out[group] = picked
		}
	}
	return out
}
