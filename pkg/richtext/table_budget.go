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
	"bytes"
	"errors"
	"strconv"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"golang.org/x/net/html"
)

// MaxTableCells counts the implied cells (spans, padded rows) of all tables in a document, since converters materialize each.
const MaxTableCells = 10_000

var ErrTooManyTableCells = errors.New("tables have too many cells")

func skipOversizedTables(doc *html.Node) {
	budget := MaxTableCells
	var visit func(n *html.Node, table *tableShape) int
	// Nested cells count towards a row's width, as in the plugin.
	visit = func(n *html.Node, table *tableShape) int {
		if n.Type == html.ElementNode && n.Data == "table" {
			if table != nil {
				// The plugin never renders a table containing another one as a grid.
				table.nested = true
			}
			own := &tableShape{rows: 1}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				visit(c, own)
			}
			cells := own.cells()
			switch {
			case own.nested:
			case cells <= budget:
				budget -= cells
			default:
				// The plugin renders these as plain rows instead of a grid.
				setAttribute(n, "role", "presentation")
			}
			return 0
		}

		width := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			width += visit(c, table)
		}
		if table == nil || n.Type != html.ElementNode {
			return width
		}

		switch n.Data {
		case "td", "th":
			colspan := spanAttribute(n, "colspan")
			rowspan := spanAttribute(n, "rowspan")
			width += colspan
			if rowspan > 1 {
				table.spannedWidth += colspan
				table.maxRowspan = max(table.maxRowspan, rowspan)
			}
		case "tr":
			table.rows++
			table.maxRowWidth = max(table.maxRowWidth, width)
		}
		return width
	}
	visit(doc, nil)
}

type tableShape struct {
	rows         int
	maxRowWidth  int
	maxRowspan   int
	spannedWidth int
	nested       bool
}

// Over-estimates once rowspans are involved: they extend the height and push later rows' cells right.
func (t *tableShape) cells() int {
	height := t.rows + t.maxRowspan
	width := t.maxRowWidth
	if t.maxRowspan > 0 {
		width = 2*t.maxRowWidth + t.spannedWidth
	}
	if height > MaxTableCells || width > MaxTableCells {
		return MaxTableCells + 1
	}
	return height * width
}

func spanAttribute(n *html.Node, key string) int {
	for _, a := range n.Attr {
		if a.Key != key {
			continue
		}
		span, err := strconv.Atoi(a.Val)
		if err != nil || span < 1 {
			return 1
		}
		return min(span, MaxTableCells+1)
	}
	return 1
}

var tableCellsKey = parser.NewContextKey()

// goldmark pads rows to the delimiter row's width and may build a table per delimiter-like line, so each counts.
type tableBudgetTransformer struct{}

func (tableBudgetTransformer) Transform(node *gast.Paragraph, reader text.Reader, pc parser.Context) {
	used, _ := pc.Get(tableCellsKey).(int)
	lines := node.Source()
	for i := 1; i < len(lines); i++ {
		line := lines[i].Bytes(reader.Source())
		if !looksLikeTableDelimiter(line) {
			continue
		}
		cols := bytes.Count(line, []byte{'|'}) + 1
		used += cols * (len(lines) - i + 1)
		if used > MaxTableCells {
			break
		}
		if i == 1 {
			// goldmark stops after a table that takes the whole paragraph.
			break
		}
	}
	pc.Set(tableCellsKey, used)
	if used > MaxTableCells {
		node.Parent().RemoveChild(node)
	}
}

func looksLikeTableDelimiter(line []byte) bool {
	hasDash := false
	for _, b := range line {
		switch b {
		case '-':
			hasDash = true
		case '|', ':', ' ', '\t', '\n', '\r', '\v', '\f':
		default:
			return false
		}
	}
	return hasDash
}

func tablesExceedBudget(pc parser.Context) bool {
	used, _ := pc.Get(tableCellsKey).(int)
	return used > MaxTableCells
}
