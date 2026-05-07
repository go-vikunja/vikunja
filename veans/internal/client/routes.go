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
// The set is intentionally tasks-centric — the bot doesn't need to manage
// users, teams, or webhooks. We grant `read_one`/`read_all` on projects so
// the bot can resolve PROJ-NN and #NN identifiers, but no project mutation.
func PermissionsForBot(routes map[string]RouteGroup) map[string][]string {
	wanted := map[string][]string{
		"tasks":           {"read_one", "read_all", "create", "update", "delete"},
		"projects":        {"read_one", "read_all"},
		"projects_views":  {"read_one", "read_all"},
		"buckets":         {"read_one", "read_all", "create", "update", "delete"},
		"labels":          {"read_one", "read_all", "create", "update", "delete"},
		"comments":        {"read_one", "read_all", "create", "update", "delete"},
		"tasks_comments":  {"read_one", "read_all", "create", "update", "delete"},
		"relations":       {"create", "delete"},
		"tasks_relations": {"create", "delete"},
		"assignees":       {"read_all", "create", "delete"},
		"tasks_assignees": {"read_all", "create", "delete"},
		"tasks_labels":    {"create", "delete", "read_all"},
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
