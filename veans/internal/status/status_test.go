package status

import (
	"testing"

	"code.vikunja.io/veans/internal/config"
)

func TestParse(t *testing.T) {
	cases := map[string]Status{
		"todo":         Todo,
		"TODO":         Todo,
		"in-progress":  InProgress,
		"in_progress":  InProgress,
		"in progress":  InProgress,
		"WIP":          InProgress,
		"doing":        InProgress,
		"in-review":    InReview,
		"review":       InReview,
		"completed":    Completed,
		"done":         Completed,
		"scrapped":     Scrapped,
		"cancelled":    Scrapped,
		"canceled":     Scrapped,
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
