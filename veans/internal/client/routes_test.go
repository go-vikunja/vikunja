package client

import "testing"

func TestPermissionsForBot_DropsUnknownGroups(t *testing.T) {
	// Server only exposes a subset of what we ask for.
	server := map[string]RouteGroup{
		"tasks": {
			"read_one":  {},
			"read_all":  {},
			"create":    {},
			"update":    {},
			// "delete" intentionally absent
		},
		"projects": {
			"read_one": {},
			"read_all": {},
		},
		// no "labels", no "comments", etc.
	}
	got := PermissionsForBot(server)

	if _, ok := got["tasks"]; !ok {
		t.Fatalf("expected tasks group in result")
	}
	for _, a := range got["tasks"] {
		if a == "delete" {
			t.Errorf("delete should have been dropped")
		}
	}
	if _, ok := got["projects"]; !ok {
		t.Fatalf("expected projects group")
	}
	if _, ok := got["labels"]; ok {
		t.Errorf("labels was not on server, should not appear in result")
	}
	if _, ok := got["nonexistent_group"]; ok {
		t.Errorf("phantom group leaked into result")
	}
}

func TestPermissionsForBot_EmptyWhenServerIsEmpty(t *testing.T) {
	got := PermissionsForBot(map[string]RouteGroup{})
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}
