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

package picker

import (
	"reflect"
	"strconv"
	"testing"

	"code.vikunja.io/veans/internal/client"
)

func proj(id, parent int64, pos float64, title string) *client.Project {
	return &client.Project{ID: id, ParentProjectID: parent, Position: pos, Title: title}
}

// titlesWithDepth flattens a forest depth-first into "title@depth" tokens.
func titlesWithDepth(forest []*node) []string {
	var out []string
	var walk func(nodes []*node)
	walk = func(nodes []*node) {
		for _, n := range nodes {
			out = append(out, n.project.Title+"@"+strconv.Itoa(n.depth))
			walk(n.children)
		}
	}
	walk(forest)
	return out
}

func TestBuildForest_SingleRoot(t *testing.T) {
	forest := buildForest([]*client.Project{proj(1, 0, 1, "Root")})
	got := titlesWithDepth(forest)
	want := []string{"Root@0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildForest_Nested(t *testing.T) {
	forest := buildForest([]*client.Project{
		proj(1, 0, 1, "Root"),
		proj(2, 1, 1, "Child"),
		proj(3, 2, 1, "Grandchild"),
	})
	got := titlesWithDepth(forest)
	want := []string{"Root@0", "Child@1", "Grandchild@2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildForest_MultipleRoots(t *testing.T) {
	forest := buildForest([]*client.Project{
		proj(1, 0, 2, "Beta"),
		proj(2, 0, 1, "Alpha"),
	})
	got := titlesWithDepth(forest)
	// Roots are sorted by position: Alpha (pos 1) before Beta (pos 2).
	want := []string{"Alpha@0", "Beta@0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildForest_SiblingOrderPositionThenTitle(t *testing.T) {
	forest := buildForest([]*client.Project{
		proj(1, 0, 0, "Root"),
		proj(2, 1, 2, "C"),
		proj(3, 1, 1, "B"),
		// same position as B — tie-break by title puts A before B.
		proj(4, 1, 1, "A"),
	})
	got := titlesWithDepth(forest)
	want := []string{"Root@0", "A@1", "B@1", "C@1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildForest_OrphanBecomesRoot(t *testing.T) {
	// Parent 99 is not in the input set — child should surface as a root.
	forest := buildForest([]*client.Project{
		proj(1, 0, 1, "Root"),
		proj(2, 99, 2, "Orphan"),
	})
	got := titlesWithDepth(forest)
	want := []string{"Root@0", "Orphan@0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildForest_DepthCorrectness(t *testing.T) {
	forest := buildForest([]*client.Project{
		proj(1, 0, 1, "A"),
		proj(2, 1, 1, "B"),
		proj(3, 2, 1, "C"),
		proj(4, 3, 1, "D"),
	})
	depthOf := map[string]int{}
	var walk func(nodes []*node)
	walk = func(nodes []*node) {
		for _, n := range nodes {
			depthOf[n.project.Title] = n.depth
			walk(n.children)
		}
	}
	walk(forest)
	for title, want := range map[string]int{"A": 0, "B": 1, "C": 2, "D": 3} {
		if depthOf[title] != want {
			t.Errorf("depth of %q = %d, want %d", title, depthOf[title], want)
		}
	}
}
