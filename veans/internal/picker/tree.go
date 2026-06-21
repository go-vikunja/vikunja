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

// Package picker renders an interactive, hierarchical, fuzzy-searchable
// project picker for `veans init`. The pure tree/flatten logic is split from
// the bubbletea TUI so it stays unit-testable.
package picker

import (
	"sort"

	"code.vikunja.io/veans/internal/client"
)

type node struct {
	project  *client.Project
	depth    int
	children []*node
}

// buildForest turns a flat project slice into a depth-annotated forest. A
// project whose ParentProjectID is absent from the input becomes a root —
// this mirrors the frontend's effective-parent behavior so children of a
// hidden or archived parent don't vanish. Siblings are ordered by Position,
// tie-broken by Title.
func buildForest(projects []*client.Project) []*node {
	byID := make(map[int64]*node, len(projects))
	for _, p := range projects {
		if p == nil {
			continue
		}
		byID[p.ID] = &node{project: p}
	}

	var roots []*node
	for _, p := range projects {
		if p == nil {
			continue
		}
		n := byID[p.ID]
		parent, ok := byID[p.ParentProjectID]
		if p.ParentProjectID == 0 || !ok {
			roots = append(roots, n)
			continue
		}
		parent.children = append(parent.children, n)
	}

	sortAndAssignDepth(roots, 0)
	return roots
}

func sortAndAssignDepth(nodes []*node, depth int) {
	sort.SliceStable(nodes, func(i, j int) bool {
		a, b := nodes[i].project, nodes[j].project
		if a.Position != b.Position {
			return a.Position < b.Position
		}
		return a.Title < b.Title
	})
	for _, n := range nodes {
		n.depth = depth
		sortAndAssignDepth(n.children, depth+1)
	}
}
