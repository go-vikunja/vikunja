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
	"fmt"
	"strings"

	"code.vikunja.io/veans/internal/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleRows = 12

var (
	dimStyle   = lipgloss.NewStyle().Faint(true)
	matchStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorMark = "❯"
)

// model is the bubbletea state for the picker. The pinned "create a new
// project" entry is the trailing row with a nil project; it is always
// selectable and never filtered out.
type model struct {
	forest []*node
	query  string
	rows   []row
	cursor int // index into rows, always on a selectable row
	offset int // first visible row index

	result    *client.Project
	createNew bool
	canceled  bool
}

func newModel(forest []*node) *model {
	m := &model{forest: forest}
	m.recompute()
	return m
}

func (m *model) recompute() {
	rows := flatten(m.forest, m.query)
	rows = append(rows, row{project: nil}) // pinned create row
	m.rows = rows
	// recompute only runs when the query changes (or on init), so snap to the
	// first match. Keeping the old cursor could leave it on the trailing create
	// row after the list narrows, making Enter create a project instead of
	// picking the visible match.
	m.cursor = 0
	m.offset = 0
	m.clampCursor()
	m.ensureVisible()
}

func (r row) isCreate() bool { return r.project == nil }

func (r row) selectable() bool { return r.isCreate() || !r.dimmed }

func (m *model) clampCursor() {
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.rows[m.cursor].selectable() {
		return
	}
	// Snap to the nearest selectable row, preferring downward.
	for i := m.cursor; i < len(m.rows); i++ {
		if m.rows[i].selectable() {
			m.cursor = i
			return
		}
	}
	for i := m.cursor; i >= 0; i-- {
		if m.rows[i].selectable() {
			m.cursor = i
			return
		}
	}
}

func (m *model) moveCursor(delta int) {
	i := m.cursor
	for {
		i += delta
		if i < 0 || i >= len(m.rows) {
			return
		}
		if m.rows[i].selectable() {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

func (m *model) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+maxVisibleRows {
		m.offset = m.cursor - maxVisibleRows + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "esc":
		m.canceled = true
		return m, tea.Quit
	case "enter":
		sel := m.rows[m.cursor]
		if sel.isCreate() {
			m.createNew = true
		} else {
			m.result = sel.project
		}
		return m, tea.Quit
	case "up":
		m.moveCursor(-1)
	case "down":
		m.moveCursor(1)
	case "backspace":
		if m.query != "" {
			r := []rune(m.query)
			m.query = string(r[:len(r)-1])
			m.recompute()
		}
	default:
		// Treat printable runes and space as query input.
		if key.Type == tea.KeyRunes || key.Type == tea.KeySpace {
			runes := key.Runes
			// KeySpace is not guaranteed to populate key.Runes; substitute a
			// literal space so multi-word fuzzy queries still work.
			if key.Type == tea.KeySpace && len(runes) == 0 {
				runes = []rune{' '}
			}
			m.query += string(runes)
			m.recompute()
		}
	}
	return m, nil
}

func (m *model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "> %s\n", m.query)

	end := min(m.offset+maxVisibleRows, len(m.rows))
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderRow(i))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "%d/%d  ↑↓ move  ⏎ pick  esc cancel\n", m.cursor+1, len(m.rows))
	return b.String()
}

func (m *model) renderRow(i int) string {
	r := m.rows[i]

	marker := "  "
	if i == m.cursor {
		marker = cursorMark + " "
	}

	indent := strings.Repeat("  ", r.depth)

	var label string
	switch {
	case r.isCreate():
		label = "Create a new project"
	case r.dimmed:
		label = dimStyle.Render(r.project.Title + projectSuffix(r.project))
	default:
		label = highlight(r.project.Title, r.matches) + dimStyle.Render(projectSuffix(r.project))
	}

	return marker + indent + label
}

// projectSuffix is the dimmed metadata appended to a project row. Titles aren't
// unique in Vikunja, so the id (and identifier when set) keeps duplicate-titled
// projects distinguishable during init.
func projectSuffix(p *client.Project) string {
	s := fmt.Sprintf("  #%d", p.ID)
	if p.Identifier != "" {
		s += " " + p.Identifier
	}
	return s
}

// highlight bolds the matched runes of title. matches are rune indexes.
func highlight(title string, matches []int) string {
	if len(matches) == 0 {
		return title
	}
	matchSet := make(map[int]bool, len(matches))
	for _, idx := range matches {
		matchSet[idx] = true
	}
	var b strings.Builder
	for i, r := range []rune(title) {
		if matchSet[i] {
			b.WriteString(matchStyle.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
