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

	"code.vikunja.io/api/pkg/web"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/util"
	"xorm.io/xorm"
)

// Neither renderer enables html.WithUnsafe() — raw HTML in the markdown stays
// inert, so the only active markup is what goldmark emits. This is what stops
// user-supplied markdown from smuggling in scripts.
var (
	gfmParser = parser.New(
		parser.WithExtensions(extension.GFMParser),
		// Right before goldmark's table transformer (priority 200).
		parser.WithParagraphTransformers(util.Prioritized[parser.ParagraphTransformer](tableBudgetTransformer{}, 199)),
	)
	gfmRenderer = html.New(html.WithExtensions(extension.GFMHTMLRenderer))

	commonMarkParser   = parser.New()
	commonMarkRenderer = html.New()
)

// CommonMarkToHTML renders plain CommonMark (no GFM extensions) to HTML.
func CommonMarkToHTML(md []byte) (string, error) {
	var buf bytes.Buffer
	if err := commonMarkRenderer.Render(&buf, md, commonMarkParser.Parse(md)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MarkdownToHTML converts GFM Markdown to canonical rich-text HTML, rewriting task
// lists into TipTap's <ul data-type="taskList"> form. Mentions are left as literal
// "@username" — see MarkdownToHTMLWithMentions to resolve them.
func MarkdownToHTML(md string) (string, error) {
	return markdownToHTML(md, nil, nil)
}

// MarkdownToHTMLWithMentions is MarkdownToHTML plus mention resolution: "@username"
// matching an existing user visible to viewer becomes a <mention-user> tag. Needs a session.
func MarkdownToHTMLWithMentions(s *xorm.Session, viewer web.Auth, md string) (string, error) {
	return markdownToHTML(md, s, viewer)
}

func markdownToHTML(md string, s *xorm.Session, viewer web.Auth) (string, error) {
	var buf bytes.Buffer
	source := []byte(md)
	pc := parser.NewContext()
	doc := gfmParser.Parse(source, parser.WithContext(pc))
	if tablesExceedBudget(pc) {
		return "", ErrTooManyTableCells
	}
	if err := gfmRenderer.Render(&buf, source, doc); err != nil {
		return "", fmt.Errorf("converting markdown to html: %w", err)
	}

	nodes, err := parseHTMLFragment(buf.Bytes())
	if err != nil {
		return "", err
	}

	for _, n := range nodes {
		convertTaskListItems(n)
	}

	if s != nil {
		if err := rebuildMentions(s, viewer, nodes); err != nil {
			return "", err
		}
	}

	out, err := renderHTMLNodes(nodes)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}
