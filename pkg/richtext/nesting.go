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
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Markdown repeats the quote prefix and list indent of every level on each line.
const maxBlockNesting = 10

func flattenDeepNesting(n *html.Node, depth int) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childDepth := depth
		if c.Type == html.ElementNode {
			switch c.Data {
			case "blockquote", "ul", "ol":
				childDepth++
				if childDepth > maxBlockNesting {
					renameToDiv(c)
				}
			case "li":
				if depth > maxBlockNesting {
					renameToDiv(c)
				}
			}
		}
		flattenDeepNesting(c, childDepth)
	}
}

func renameToDiv(n *html.Node) {
	n.Data = "div"
	n.DataAtom = atom.Div
}
