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
	"strings"
	"testing"
)

func TestComposeDescription_FullReplace(t *testing.T) {
	f := &updateFlags{description: "new body", descriptionIsSet: true}
	got, changed, err := composeDescription("old body", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "new body" {
		t.Fatalf("got %q changed=%v", got, changed)
	}
}

func TestComposeDescription_SurgicalReplace(t *testing.T) {
	f := &updateFlags{
		replaceOld: "TODO",
		replaceNew: "DONE",
	}
	got, changed, err := composeDescription("- [ ] TODO part 1\n- [ ] something else", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !strings.Contains(got, "DONE part 1") {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_ReplaceNotUnique(t *testing.T) {
	f := &updateFlags{
		replaceOld: "x",
		replaceNew: "y",
	}
	if _, _, err := composeDescription("xxx", f); err == nil {
		t.Fatal("expected error on non-unique match")
	}
}

func TestComposeDescription_ReplaceNotFound(t *testing.T) {
	f := &updateFlags{
		replaceOld: "missing",
		replaceNew: "y",
	}
	if _, _, err := composeDescription("hello", f); err == nil {
		t.Fatal("expected error on no match")
	}
}

func TestComposeDescription_Append(t *testing.T) {
	f := &updateFlags{descriptionApp: "## Notes"}
	got, changed, err := composeDescription("body", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "body\n## Notes" {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_AppendOnEmpty(t *testing.T) {
	f := &updateFlags{descriptionApp: "first line"}
	got, changed, err := composeDescription("", f)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || got != "first line" {
		t.Fatalf("got %q", got)
	}
}

func TestComposeDescription_NoOp(t *testing.T) {
	f := &updateFlags{}
	got, changed, err := composeDescription("body", f)
	if err != nil {
		t.Fatal(err)
	}
	if changed || got != "body" {
		t.Fatalf("expected no-op, got %q changed=%v", got, changed)
	}
}

func TestComposeDescription_ReplaceNewWithoutOld(t *testing.T) {
	f := &updateFlags{replaceNew: "y"}
	if _, _, err := composeDescription("body", f); err == nil {
		t.Fatal("expected error: --description-replace-new without --description-replace-old")
	}
}

func TestNormalizeLabelTitle(t *testing.T) {
	cases := map[string]string{
		"foo":                    "veans:foo",
		"veans:bar":              "veans:bar",
		"  baz  ":                "veans:baz",
		"veans:already-prefixed": "veans:already-prefixed",
	}
	for in, want := range cases {
		if got := normalizeLabelTitle(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
