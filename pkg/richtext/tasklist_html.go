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
	"fmt"
	"strings"

	"github.com/JohannesKaufmann/dom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// parseHTMLFragment parses an HTML fragment in a <body> context (so tables/lists parse).
func parseHTMLFragment(in []byte) ([]*html.Node, error) {
	context := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(bytes.NewReader(in), context)
	if err != nil {
		return nil, fmt.Errorf("parsing converted html: %w", err)
	}
	return nodes, nil
}

func renderHTMLNodes(nodes []*html.Node) (string, error) {
	var buf bytes.Buffer
	for _, n := range nodes {
		if err := html.Render(&buf, n); err != nil {
			return "", fmt.Errorf("rendering converted html: %w", err)
		}
	}
	return buf.String(), nil
}

// convertTaskListItems rewrites goldmark's GFM task-list output
// (<li><input type="checkbox"> text</li>) into the TipTap
// <ul data-type="taskList"><li data-type="taskItem" data-checked="…"><p>text</p></li>
// shape the web editor and resetDescriptionChecklist (recurring-task reset) expect.
func convertTaskListItems(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertTaskListItems(c)
	}

	if n.Type != html.ElementNode || n.Data != "li" {
		return
	}

	input := leadingCheckbox(n)
	if input == nil {
		return
	}

	_, checked := dom.GetAttribute(input, "checked")
	dom.RemoveNode(input)

	setAttribute(n, "data-type", "taskItem")
	setAttribute(n, "data-checked", boolString(checked))
	wrapLeadingInlineInParagraph(n)

	if p := n.Parent; p != nil && p.Type == html.ElementNode && (p.Data == "ul" || p.Data == "ol") {
		setAttribute(p, "data-type", "taskList")
	}
}

// leadingCheckbox returns the <input type="checkbox"> at the start of li (after
// skipping insignificant whitespace), or nil if li isn't a task item.
func leadingCheckbox(li *html.Node) *html.Node {
	for c := li.FirstChild; c != nil; c = c.NextSibling {
		if isWhitespaceText(c) {
			continue
		}
		if c.Type == html.ElementNode && c.Data == "input" && dom.GetAttributeOr(c, "type", "") == "checkbox" {
			return c
		}
		return nil
	}
	return nil
}

// wrapLeadingInlineInParagraph moves li's leading inline content (everything up
// to the first nested list) into a <p>, matching TipTap's taskItem shape.
func wrapLeadingInlineInParagraph(li *html.Node) {
	var inline []*html.Node
	for c := li.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
			break
		}
		inline = append(inline, c)
	}

	allWhitespace := true
	for _, c := range inline {
		if !isWhitespaceText(c) {
			allWhitespace = false
			break
		}
	}
	if len(inline) == 0 || allWhitespace {
		return
	}

	p := &html.Node{Type: html.ElementNode, Data: "p", DataAtom: atom.P}
	for _, c := range inline {
		li.RemoveChild(c)
		p.AppendChild(c)
	}
	li.InsertBefore(p, li.FirstChild)
	trimEdgeWhitespace(p)
}

// trimEdgeWhitespace trims leading/trailing whitespace from the first and last
// text nodes of n so the wrapped paragraph doesn't keep goldmark's "<input> "
// spacing or trailing newline.
func trimEdgeWhitespace(n *html.Node) {
	if first := n.FirstChild; first != nil && first.Type == html.TextNode {
		first.Data = strings.TrimLeft(first.Data, " \t\n\r")
	}
	if last := n.LastChild; last != nil && last.Type == html.TextNode {
		last.Data = strings.TrimRight(last.Data, " \t\n\r")
	}
}

func setAttribute(n *html.Node, key, val string) {
	for i, a := range n.Attr {
		if a.Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func isWhitespaceText(n *html.Node) bool {
	return n.Type == html.TextNode && strings.TrimSpace(n.Data) == ""
}
