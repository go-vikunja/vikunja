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
	"unicode/utf8"

	"code.vikunja.io/veans/internal/client"
	"github.com/sahilm/fuzzy"
)

// row is one visible line in the picker. matches holds rune indexes into the
// title for highlighting; dimmed rows are kept only as context for a matching
// descendant and are skipped by the cursor.
type row struct {
	project *client.Project
	depth   int
	dimmed  bool
	matches []int
}

// flatten walks the forest depth-first into a render list. An empty query
// returns every node undimmed. A non-empty query fuzzy-matches each title
// (case-insensitive, via sahilm/fuzzy) and keeps a node iff it matches or any
// descendant is kept; a kept-but-non-matching node is dimmed context.
func flatten(forest []*node, query string) []row {
	if query == "" {
		var rows []row
		var walk func(nodes []*node)
		walk = func(nodes []*node) {
			for _, n := range nodes {
				rows = append(rows, row{project: n.project, depth: n.depth})
				walk(n.children)
			}
		}
		walk(forest)
		return rows
	}

	var rows []row
	var walk func(n *node) bool
	walk = func(n *node) bool {
		matches, matched := matchTitle(query, n.project.Title)

		start := len(rows)
		rows = append(rows, row{}) // placeholder; finalized only if kept

		descendantKept := false
		for _, c := range n.children {
			if walk(c) {
				descendantKept = true
			}
		}

		if !matched && !descendantKept {
			rows = rows[:start]
			return false
		}
		rows[start] = row{
			project: n.project,
			depth:   n.depth,
			dimmed:  !matched,
			matches: matches,
		}
		return true
	}

	for _, n := range forest {
		walk(n)
	}
	return rows
}

// matchTitle reports whether title fuzzy-matches query and, if so, the rune
// indexes of the matched characters. sahilm/fuzzy reports byte indexes, so we
// translate them to rune offsets for correct highlighting of multibyte titles.
func matchTitle(query, title string) (runeMatches []int, matched bool) {
	results := fuzzy.Find(query, []string{title})
	if len(results) == 0 {
		return nil, false
	}
	return byteToRuneIndexes(title, results[0].MatchedIndexes), true
}

func byteToRuneIndexes(s string, byteIdx []int) []int {
	if len(byteIdx) == 0 {
		return nil
	}
	want := make(map[int]bool, len(byteIdx))
	for _, b := range byteIdx {
		want[b] = true
	}
	out := make([]int, 0, len(byteIdx))
	runePos := 0
	for b := 0; b < len(s); {
		if want[b] {
			out = append(out, runePos)
		}
		_, size := utf8.DecodeRuneInString(s[b:])
		b += size
		runePos++
	}
	return out
}
