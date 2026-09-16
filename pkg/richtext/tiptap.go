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
	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
)

// registerTipTapRules teaches the HTML→Markdown converter about the two
// Vikunja-specific nodes that standard GFM doesn't model: TipTap mentions and
// TipTap task lists.
func registerTipTapRules(conv *converter.Converter) {
	// Empty mention elements (the common stored form is <mention-user data-id data-label></mention-user>)
	// would otherwise be treated as content-less by the whitespace collapser, eating the
	// following space. Giving them a text child before collapse (PriorityLate) preserves it.
	conv.Register.PreRenderer(ensureMentionContent, converter.PriorityEarly)
	conv.Register.RendererFor("mention-user", converter.TagTypeInline, renderMentionUser, converter.PriorityEarly)

	// Normalize TipTap task-list items to a single <input type="checkbox"> that
	// renderTaskCheckbox turns into the GFM "[x]"/"[ ]" marker. We drive off the
	// <li data-checked> attribute (the same source of truth resetDescriptionChecklist
	// uses) rather than TipTap's <label><input> chrome, which may not always be present.
	conv.Register.PreRenderer(normalizeTaskListItems, converter.PriorityEarly)
	conv.Register.RendererFor("input", converter.TagTypeInline, renderTaskCheckbox, converter.PriorityEarly)
}

// renderMentionUser converts <mention-user data-id="username"> to "@username"
// (label and inner text dropped). Tags without data-id fall through to the
// default renderer, keeping their inner text.
func renderMentionUser(_ converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	username := dom.GetAttributeOr(n, "data-id", "")
	if username == "" {
		return converter.RenderTryNext
	}

	// Written directly to the writer so the username isn't markdown-escaped;
	// the inbound side re-tokenizes "@username" verbatim. The writer is
	// buffer-backed and never errors.
	_, _ = w.WriteString("@" + username)
	return converter.RenderSuccess
}

// ensureMentionContent gives every mention with a data-id a text child if it has
// none, so the whitespace collapser keeps it (and the surrounding spaces). The
// child is never rendered — renderMentionUser writes "@data-id" and stops.
func ensureMentionContent(_ converter.Context, doc *html.Node) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "mention-user" && n.FirstChild == nil {
			if username := dom.GetAttributeOr(n, "data-id", ""); username != "" {
				n.AppendChild(&html.Node{Type: html.TextNode, Data: "@" + username})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

// renderTaskCheckbox emits the GFM task-list marker for the normalized checkbox
// input. The trailing space separates it from the item text ("- [x] text").
func renderTaskCheckbox(_ converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	if dom.GetAttributeOr(n, "type", "") != "checkbox" {
		return converter.RenderTryNext
	}

	marker := "[ ] "
	if _, checked := dom.GetAttribute(n, "checked"); checked {
		marker = "[x] "
	}
	_, _ = w.WriteString(marker)
	return converter.RenderSuccess
}

// normalizeTaskListItems rewrites every <li data-checked="…"> so its checkbox
// state is carried by a single leading <input type="checkbox">, removing
// TipTap's <label> chrome. This makes the marker independent of whether the
// stored HTML used the full TipTap form or the bare data-checked form.
func normalizeTaskListItems(_ converter.Context, doc *html.Node) {
	var items []*html.Node
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			if _, ok := dom.GetAttribute(n, "data-checked"); ok {
				items = append(items, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collect(c)
		}
	}
	collect(doc)

	for _, li := range items {
		checked := dom.GetAttributeOr(li, "data-checked", "false") == "true"

		// Drop the existing checkbox chrome (<label><input><span>) so we don't
		// render a duplicate or stale marker.
		for _, child := range dom.AllChildNodes(li) {
			if child.Type == html.ElementNode && (child.Data == "label" || child.Data == "input") {
				dom.RemoveNode(child)
			}
		}

		input := &html.Node{
			Type: html.ElementNode,
			Data: "input",
			Attr: []html.Attribute{{Key: "type", Val: "checkbox"}},
		}
		if checked {
			input.Attr = append(input.Attr, html.Attribute{Key: "checked", Val: "checked"})
		}

		// Insert the marker inside the item's first paragraph so it stays inline
		// with the text ("- [x] text"). TipTap wraps task text in <div><p>…</p></div>;
		// inserting at the <li> level instead would put a block boundary between
		// the marker and the text.
		host := firstParagraph(li)
		if host == nil {
			host = li
		}
		host.InsertBefore(input, host.FirstChild)
	}
}

func firstParagraph(n *html.Node) *html.Node {
	for _, c := range dom.AllChildNodes(n) {
		if c.Type == html.ElementNode && c.Data == "p" {
			return c
		}
		if found := firstParagraph(c); found != nil {
			return found
		}
	}
	return nil
}
