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
// Order matches the columns we want left-to-right in a Kanban view.
var CanonicalBucketTitles = []string{
	"Todo",
	"In Progress",
	"In Review",
	"Done",
	"Scrapped",
}

// BucketTitleAliases lists titles that count as the canonical bucket for
// each status. Vikunja's default Kanban view ships with "To-Do", "Doing"
// and "Done" buckets — we accept those so a vanilla project doesn't grow
// parallel buckets when veans init runs against it. The first entry is
// always the canonical name used by CanonicalBucketTitles.
var BucketTitleAliases = map[Status][]string{
	Todo:       {"Todo", "To-Do", "ToDo", "To Do", "To do", "Backlog"},
	InProgress: {"In Progress", "In-Progress", "Doing", "WIP", "In progress"},
	InReview:   {"In Review", "In-Review", "Review", "In review"},
	Completed:  {"Done", "Completed", "Complete"},
	Scrapped:   {"Scrapped", "Cancelled", "Canceled", "Won't Do", "Wontfix"},
}

// MatchBucketTitle reports whether `title` matches `s` either as the
// canonical title or one of its aliases. Comparison is case-insensitive
// and tolerant of stray whitespace.
func MatchBucketTitle(s Status, title string) bool {
	want := normalizeBucketTitle(title)
	for _, alias := range BucketTitleAliases[s] {
		if normalizeBucketTitle(alias) == want {
			return true
		}
	}
	return false
}

func normalizeBucketTitle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
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
