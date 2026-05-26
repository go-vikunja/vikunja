// Package status maps the five canonical veans statuses to Vikunja bucket
// IDs and the `done` flag. The mapping is canonical and reflected verbatim
// in the agent prompt (see internal/commands/prompt.tmpl).
package status

import (
	"fmt"
	"strings"

	"code.vikunja.io/veans/internal/config"
)

// Status is the agent-facing state name.
type Status string

const (
	Todo       Status = "todo"
	InProgress Status = "in-progress"
	InReview   Status = "in-review"
	Completed  Status = "completed"
	Scrapped   Status = "scrapped"
)

// All returns the canonical statuses in display order.
func All() []Status {
	return []Status{Todo, InProgress, InReview, Completed, Scrapped}
}

// CanonicalBucketTitles is the strict-with-override list seeded by `init`.
var CanonicalBucketTitles = []string{
	"Todo",
	"In Progress",
	"In Review",
	"Done",
	"Scrapped",
}

// BucketTitle returns the bucket name that backs each status.
func (s Status) BucketTitle() string {
	switch s {
	case Todo:
		return "Todo"
	case InProgress:
		return "In Progress"
	case InReview:
		return "In Review"
	case Completed:
		return "Done"
	case Scrapped:
		return "Scrapped"
	}
	return ""
}

// Done reports whether tasks in this status should have done=true.
func (s Status) Done() bool {
	return s == Completed || s == Scrapped
}

// Parse normalizes user input. Accepts the canonical hyphenated form, plus
// underscored/snake variants and a couple of natural-language synonyms.
func Parse(raw string) (Status, error) {
	n := strings.TrimSpace(strings.ToLower(raw))
	n = strings.ReplaceAll(n, "_", "-")
	n = strings.ReplaceAll(n, " ", "-")
	switch n {
	case "todo":
		return Todo, nil
	case "in-progress", "wip", "doing":
		return InProgress, nil
	case "in-review", "review":
		return InReview, nil
	case "completed", "done":
		return Completed, nil
	case "scrapped", "cancelled", "canceled":
		return Scrapped, nil
	}
	return "", fmt.Errorf("unknown status %q (expected one of: %s)",
		raw, strings.Join(allStrings(), ", "))
}

// BucketID resolves a status to the bucket ID stored in .veans.yml.
func BucketID(s Status, b config.Buckets) (int64, error) {
	switch s {
	case Todo:
		return b.Todo, nil
	case InProgress:
		return b.InProgress, nil
	case InReview:
		return b.InReview, nil
	case Completed:
		return b.Done, nil
	case Scrapped:
		return b.Scrapped, nil
	}
	return 0, fmt.Errorf("unknown status %q", s)
}

// FromBucketID is the inverse of BucketID — used by `list` to render the
// status of a task fetched from the API.
func FromBucketID(id int64, b config.Buckets) Status {
	switch id {
	case b.Todo:
		return Todo
	case b.InProgress:
		return InProgress
	case b.InReview:
		return InReview
	case b.Done:
		return Completed
	case b.Scrapped:
		return Scrapped
	}
	return ""
}

func allStrings() []string {
	out := make([]string, 0, 5)
	for _, s := range All() {
		out = append(out, string(s))
	}
	return out
}
