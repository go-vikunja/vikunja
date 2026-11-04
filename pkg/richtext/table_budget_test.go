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

package richtext

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLToMarkdownTableBudget(t *testing.T) {
	t.Run("huge spans render without expanding", func(t *testing.T) {
		got, err := HTMLToMarkdown(`<table><tr><td rowspan="500" colspan="500">x</td></tr></table>`)
		require.NoError(t, err)
		assert.Contains(t, got, "x")
		assert.Less(t, len(got), 100)
	})

	t.Run("huge colspan renders without expanding", func(t *testing.T) {
		got, err := HTMLToMarkdown(`<table><tr><td colspan="20000">x</td></tr></table>`)
		require.NoError(t, err)
		assert.Less(t, len(got), 100)
	})

	t.Run("overflowing span values render without expanding", func(t *testing.T) {
		got, err := HTMLToMarkdown(`<table><tr><td rowspan="9223372036854775807" colspan="9223372036854775807">x</td></tr></table>`)
		require.NoError(t, err)
		assert.Less(t, len(got), 100)
	})

	t.Run("one wide row and many short rows render bounded", func(t *testing.T) {
		var b strings.Builder
		b.WriteString("<table><tr>")
		b.WriteString(strings.Repeat("<td>a</td>", 500))
		b.WriteString("</tr>")
		b.WriteString(strings.Repeat("<tr><td>b</td></tr>", 500))
		b.WriteString("</table>")

		got, err := HTMLToMarkdown(b.String())
		require.NoError(t, err)
		// Expanded as a grid this would be 500x500 padded cells.
		assert.Less(t, len(got), b.Len())
	})

	t.Run("budget is shared by all tables of a document", func(t *testing.T) {
		table := "<table><tr>" + strings.Repeat("<td>a</td>", 50) + "</tr>" +
			strings.Repeat("<tr><td>b</td></tr>", 50) + "</table>"
		input := strings.Repeat(table, 200)

		got, err := HTMLToMarkdown(input)
		require.NoError(t, err)
		assert.Less(t, len(got), len(input))
	})

	t.Run("large span-free table within the budget survives a round trip", func(t *testing.T) {
		// 601 rows x 10 columns: within MaxTableCells, but over half of it.
		md := strings.Repeat("| a ", 10) + "|\n" + strings.Repeat("|---", 10) + "|" + strings.Repeat("\n"+strings.Repeat("| 1 ", 10)+"|", 600)

		rendered, err := MarkdownToHTML(md)
		require.NoError(t, err)
		got, err := HTMLToMarkdown(rendered)
		require.NoError(t, err)
		assert.Equal(t, md, got)
	})

	t.Run("normal table with spans still renders as a table", func(t *testing.T) {
		got, err := HTMLToMarkdown(`<table><tr><th colspan="2">a</th></tr><tr><td>1</td><td>2</td></tr></table>`)
		require.NoError(t, err)
		assert.Equal(t, "| a |   |\n|---|---|\n| 1 | 2 |", got)
	})
}

func TestMarkdownToHTMLTableBudget(t *testing.T) {
	t.Run("wide header with many rows is rejected", func(t *testing.T) {
		md := strings.Repeat("|a", 1000) + "|\n" + strings.Repeat("|-", 1000) + "|\n" + strings.Repeat("x\n", 1000)
		_, err := MarkdownToHTML(md)
		require.ErrorIs(t, err, ErrTooManyTableCells)
	})

	t.Run("repeated delimiter rows are rejected", func(t *testing.T) {
		md := "x\ny\n" + strings.Repeat("|-|\n", 1000)
		_, err := MarkdownToHTML(md)
		require.ErrorIs(t, err, ErrTooManyTableCells)
	})

	t.Run("tables in a blockquote count", func(t *testing.T) {
		md := "> " + strings.Repeat("|a", 1000) + "|\n> " + strings.Repeat("|-", 1000) + "|\n" + strings.Repeat("> x\n", 1000)
		_, err := MarkdownToHTML(md)
		require.ErrorIs(t, err, ErrTooManyTableCells)
	})

	t.Run("budget is shared by all tables of a document", func(t *testing.T) {
		table := strings.Repeat("|a", 50) + "|\n" + strings.Repeat("|-", 50) + "|\n" + strings.Repeat("x\n", 50)
		_, err := MarkdownToHTML(strings.Repeat(table+"\n", 10))
		require.ErrorIs(t, err, ErrTooManyTableCells)
	})

	t.Run("normal table converts", func(t *testing.T) {
		table := strings.Repeat("|a", 20) + "|\n" + strings.Repeat("|-", 20) + "|\n" + strings.Repeat("|1", 20) + "|\n"
		got, err := MarkdownToHTML(strings.Repeat(table+"\n", 5))
		require.NoError(t, err)
		assert.Contains(t, got, "<table>")
	})
}
