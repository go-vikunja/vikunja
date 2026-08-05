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

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"xorm.io/xorm"
)

// markdownConverter renders GFM but never enables html.WithUnsafe() — raw HTML in
// the markdown stays inert, so the only active markup is what goldmark emits. This
// is what stops user-supplied markdown from smuggling in scripts.
var markdownConverter = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// MarkdownToHTML converts GFM Markdown to canonical rich-text HTML, rewriting task
// lists into TipTap's <ul data-type="taskList"> form. Mentions are left as literal
// "@username" — see MarkdownToHTMLWithMentions to resolve them.
func MarkdownToHTML(md string) (string, error) {
	return markdownToHTML(md, nil)
}

// MarkdownToHTMLWithMentions is MarkdownToHTML plus mention resolution: "@username"
// matching an existing user becomes a <mention-user> tag. Needs a session.
func MarkdownToHTMLWithMentions(s *xorm.Session, md string) (string, error) {
	return markdownToHTML(md, s)
}

func markdownToHTML(md string, s *xorm.Session) (string, error) {
	var buf bytes.Buffer
	if err := markdownConverter.Convert([]byte(md), &buf); err != nil {
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
		if err := rebuildMentions(s, nodes); err != nil {
			return "", err
		}
	}

	out, err := renderHTMLNodes(nodes)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}
