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
