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

// Package richtext converts Vikunja's canonical rich-text HTML to and from
// Markdown at the API/CalDAV boundaries. Storage stays HTML; only the wire
// representation changes.
package richtext

import (
	"fmt"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// HTMLToMarkdown converts rich-text HTML to GFM Markdown. Trimmed, so an empty
// document ("<p></p>") yields "".
func HTMLToMarkdown(htmlInput string) (string, error) {
	md, err := newHTMLToMarkdownConverter().ConvertString(htmlInput)
	if err != nil {
		return "", fmt.Errorf("converting html to markdown: %w", err)
	}

	return strings.TrimSpace(md), nil
}

// newHTMLToMarkdownConverter builds a GFM converter. Per call: the registered
// handlers aren't safe for concurrent reuse, and conversion is cheap.
func newHTMLToMarkdownConverter() *converter.Converter {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
			strikethrough.NewStrikethroughPlugin(),
		),
	)

	registerTipTapRules(conv)

	return conv
}
