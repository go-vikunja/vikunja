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

package status

import (
	"testing"

	"code.vikunja.io/veans/internal/config"
)

func TestParse(t *testing.T) {
	cases := map[string]Status{
		"todo":        Todo,
		"TODO":        Todo,
		"in-progress": InProgress,
		"in_progress": InProgress,
		"in progress": InProgress,
		"WIP":         InProgress,
		"doing":       InProgress,
		"in-review":   InReview,
		"review":      InReview,
		"completed":   Completed,
		"done":        Completed,
		"scrapped":    Scrapped,
		"cancelled":   Scrapped,
		"canceled":    Scrapped,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("Parse(%q): got %q, want %q", in, got, want)
		}
	}
	if _, err := Parse("nope"); err == nil {
		t.Errorf("Parse(\"nope\"): expected error")
	}
}

func TestDoneFlag(t *testing.T) {
	if !Completed.Done() || !Scrapped.Done() {
		t.Fatal("Completed/Scrapped should be done")
	}
	if Todo.Done() || InProgress.Done() || InReview.Done() {
		t.Fatal("Todo/InProgress/InReview should not be done")
	}
}

func TestMatchBucketTitle(t *testing.T) {
	cases := []struct {
		title string
		want  Status
	}{
		// Vikunja defaults
		{"To-Do", Todo},
		{"Doing", InProgress},
		{"Done", Completed},
		// Canonical titles
		{"Todo", Todo},
		{"In Progress", InProgress},
		{"In Review", InReview},
		{"Scrapped", Scrapped},
		// Case-insensitive + whitespace tolerant
		{"  todo  ", Todo},
		{"DOING", InProgress},
		// A few common aliases
		{"WIP", InProgress},
		{"Backlog", Todo},
		{"Cancelled", Scrapped},
		{"Won't Do", Scrapped},
	}
	for _, c := range cases {
		matched := false
		for _, s := range All() {
			if MatchBucketTitle(s, c.title) {
				if s != c.want {
					t.Errorf("MatchBucketTitle(%q): matched %q, want %q", c.title, s, c.want)
				}
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("MatchBucketTitle(%q): no status matched, want %q", c.title, c.want)
		}
	}

	// Negative: a non-canonical name shouldn't match anything.
	for _, s := range All() {
		if MatchBucketTitle(s, "random-bucket-name") {
			t.Errorf("MatchBucketTitle(%q, \"random\") unexpectedly true", s)
		}
	}
}

func TestBucketIDRoundTrip(t *testing.T) {
	b := config.Buckets{Todo: 11, InProgress: 12, InReview: 13, Done: 14, Scrapped: 15}
	for _, s := range All() {
		id, err := BucketID(s, b)
		if err != nil {
			t.Fatalf("BucketID(%q): %v", s, err)
		}
		if got := FromBucketID(id, b); got != s {
			t.Errorf("FromBucketID(%d) = %q, want %q", id, got, s)
		}
	}
}
