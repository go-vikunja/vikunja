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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLToMarkdown(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "heading",
			html: "<h1>Title</h1>",
			want: "# Title",
		},
		{
			name: "bold and italic",
			html: "<p><strong>bold</strong> and <em>italic</em></p>",
			want: "**bold** and *italic*",
		},
		{
			name: "link",
			html: `<p>See <a href="https://vikunja.io">the site</a></p>`,
			want: "See [the site](https://vikunja.io)",
		},
		{
			name: "inline code",
			html: "<p>run <code>mage build</code> first</p>",
			want: "run `mage build` first",
		},
		{
			name: "fenced code block keeps language",
			html: `<pre><code class="language-go">fmt.Println("hi")</code></pre>`,
			want: "```go\nfmt.Println(\"hi\")\n```",
		},
		{
			name: "blockquote",
			html: "<blockquote><p>quoted text</p></blockquote>",
			want: "> quoted text",
		},
		{
			name: "unordered list",
			html: "<ul><li>one</li><li>two</li></ul>",
			want: "- one\n- two",
		},
		{
			name: "ordered list",
			html: "<ol><li>one</li><li>two</li></ol>",
			want: "1. one\n2. two",
		},
		{
			name: "nested list",
			html: "<ul><li>one<ul><li>nested</li></ul></li><li>two</li></ul>",
			want: "- one\n  \n  - nested\n- two",
		},
		{
			name: "gfm table",
			html: "<table><thead><tr><th>a</th><th>b</th></tr></thead><tbody><tr><td>1</td><td>2</td></tr></tbody></table>",
			want: "| a | b |\n|---|---|\n| 1 | 2 |",
		},
		{
			name: "strikethrough",
			html: "<p><del>gone</del></p>",
			want: "~~gone~~",
		},
		{
			name: "empty paragraph is empty string",
			html: "<p></p>",
			want: "",
		},
		{
			name: "whitespace only is empty string",
			html: "<p>   </p>",
			want: "",
		},
		{
			name: "unknown element degrades without leaking tags",
			html: "<p>hello <unknowntag>world</unknowntag></p>",
			want: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLToMarkdown(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
