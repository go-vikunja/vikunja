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
	"testing"

	"code.vikunja.io/veans/internal/client"
)

func sampleForest() []*node {
	return buildForest([]*client.Project{
		proj(1, 0, 1, "Backend"),
		proj(2, 1, 1, "Frontend"),
		proj(3, 1, 2, "Database"),
		proj(4, 0, 2, "Marketing"),
	})
}

func rowTitles(rows []row) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.project.Title
	}
	return out
}

func TestFlatten_EmptyQuery(t *testing.T) {
	rows := flatten(sampleForest(), "")
	wantTitles := []string{"Backend", "Frontend", "Database", "Marketing"}
	if got := rowTitles(rows); !reflect.DeepEqual(got, wantTitles) {
		t.Fatalf("titles: got %v, want %v", got, wantTitles)
	}
	wantDepths := []int{0, 1, 1, 0}
	for i, r := range rows {
		if r.depth != wantDepths[i] {
			t.Errorf("row %d depth = %d, want %d", i, r.depth, wantDepths[i])
		}
		if r.dimmed {
			t.Errorf("row %d should not be dimmed on empty query", i)
		}
		if r.matches != nil {
			t.Errorf("row %d should have nil matches on empty query", i)
		}
	}
}

func TestFlatten_DeepChildSurfacesDimmedAncestor(t *testing.T) {
	// "Frontend" is a child of "Backend"; matching it must keep "Backend"
	// as a dimmed context row.
	rows := flatten(sampleForest(), "frontend")
	wantTitles := []string{"Backend", "Frontend"}
	if got := rowTitles(rows); !reflect.DeepEqual(got, wantTitles) {
		t.Fatalf("titles: got %v, want %v", got, wantTitles)
	}
	if !rows[0].dimmed {
		t.Error("ancestor Backend should be dimmed (context only)")
	}
	if rows[1].dimmed {
		t.Error("matching Frontend should not be dimmed")
	}
}

func TestFlatten_MatchingNodeCarriesMatchIndexes(t *testing.T) {
	rows := flatten(sampleForest(), "front")
	var frontend *row
	for i := range rows {
		if rows[i].project.Title == "Frontend" {
			frontend = &rows[i]
		}
	}
	if frontend == nil {
		t.Fatal("Frontend row missing")
	}
	// "front" should match the leading runes of "Frontend".
	want := []int{0, 1, 2, 3, 4}
	if !reflect.DeepEqual(frontend.matches, want) {
		t.Fatalf("matches: got %v, want %v", frontend.matches, want)
	}
}

func TestFlatten_NonMatchingSiblingsDropped(t *testing.T) {
	// Matching "Marketing" must not pull in "Backend"/"Frontend"/"Database".
	rows := flatten(sampleForest(), "marketing")
	wantTitles := []string{"Marketing"}
	if got := rowTitles(rows); !reflect.DeepEqual(got, wantTitles) {
		t.Fatalf("titles: got %v, want %v", got, wantTitles)
	}
}

func TestFlatten_NoMatchYieldsEmpty(t *testing.T) {
	rows := flatten(sampleForest(), "zzzzz")
	if len(rows) != 0 {
		t.Fatalf("expected no rows, got %v", rowTitles(rows))
	}
}

func TestFlatten_CaseInsensitive(t *testing.T) {
	lower := flatten(sampleForest(), "backend")
	upper := flatten(sampleForest(), "BACKEND")
	if !reflect.DeepEqual(rowTitles(lower), rowTitles(upper)) {
		t.Fatalf("case sensitivity differs: %v vs %v", rowTitles(lower), rowTitles(upper))
	}
	if len(lower) == 0 {
		t.Fatal("expected at least one match for 'backend'")
	}
}
