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

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want string
	}{
		{
			name: "heading and bold",
			md:   "# Title\n\nsome **bold** text",
			want: "<h1>Title</h1>\n<p>some <strong>bold</strong> text</p>",
		},
		{
			name: "link",
			md:   "see [the site](https://vikunja.io)",
			want: `<p>see <a href="https://vikunja.io">the site</a></p>`,
		},
		{
			name: "task list becomes tiptap dom",
			md:   "- [x] done\n- [ ] todo",
			want: "<ul data-type=\"taskList\">\n<li data-type=\"taskItem\" data-checked=\"true\"><p>done</p></li>\n<li data-type=\"taskItem\" data-checked=\"false\"><p>todo</p></li>\n</ul>",
		},
		{
			name: "nested task list",
			md:   "- [ ] parent\n  - [x] child",
			want: "<ul data-type=\"taskList\">\n<li data-type=\"taskItem\" data-checked=\"false\"><p>parent</p><ul data-type=\"taskList\">\n<li data-type=\"taskItem\" data-checked=\"true\"><p>child</p></li>\n</ul>\n</li>\n</ul>",
		},
		{
			name: "task list keeps inline formatting",
			md:   "- [x] task with **bold** and a [link](https://x.io)",
			want: "<ul data-type=\"taskList\">\n<li data-type=\"taskItem\" data-checked=\"true\"><p>task with <strong>bold</strong> and a <a href=\"https://x.io\">link</a></p></li>\n</ul>",
		},
		{
			name: "plain list is not a task list",
			md:   "- one\n- two",
			want: "<ul>\n<li>one</li>\n<li>two</li>\n</ul>",
		},
		{
			name: "pipe table",
			md:   "| a | b |\n|---|---|\n| 1 | 2 |",
			want: "<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>",
		},
		{
			name: "strikethrough",
			md:   "~~gone~~",
			want: "<p><del>gone</del></p>",
		},
		{
			name: "empty markdown is empty",
			md:   "",
			want: "",
		},
		{
			name: "whitespace markdown is empty",
			md:   "   \n  ",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownToHTML(tt.md)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMarkdownToHTML_NoUnsafe proves goldmark runs without html.WithUnsafe():
// raw HTML in the markdown must never become active markup.
func TestMarkdownToHTML_NoUnsafe(t *testing.T) {
	got, err := MarkdownToHTML("text with <script>alert(1)</script> raw html")
	require.NoError(t, err)
	assert.NotContains(t, got, "<script>")
	assert.NotContains(t, got, "</script>")
}
