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

package commands

import (
	"context"
	"strings"

	"code.vikunja.io/veans/internal/client"
)

// labelNamespace is auto-prepended to label names that don't already have it,
// so the agent's labels live in their own corner of the user's global label
// list and don't pollute manually-curated labels.
const labelNamespace = "veans:"

func normalizeLabelTitle(raw string) string {
	t := strings.TrimSpace(raw)
	if t == "" {
		return ""
	}
	if strings.HasPrefix(t, labelNamespace) {
		return t
	}
	return labelNamespace + t
}

// getOrCreateLabelByTitle returns the ID of the label with the given title,
// creating it under the current user if it doesn't exist. Labels are global
// per user in Vikunja, so this only finds labels visible to whoever the
// `c` client is authenticated as (i.e. the bot when called from veans).
func getOrCreateLabelByTitle(ctx context.Context, c *client.Client, title string) (*client.Label, error) {
	existing, err := c.ListLabels(ctx, title)
	if err != nil {
		return nil, err
	}
	for _, l := range existing {
		if l.Title == title {
			return l, nil
		}
	}
	created, err := c.CreateLabel(ctx, &client.Label{Title: title})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// findLabelOnTask returns the label with the given (already-normalized)
// title attached to the task, or nil. Used by --label-remove to know which
// label ID to detach.
func findLabelOnTask(t *client.Task, title string) *client.Label {
	for _, l := range t.Labels {
		if l != nil && l.Title == title {
			return l
		}
	}
	return nil
}
