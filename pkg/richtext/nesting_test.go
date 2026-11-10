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

func TestHTMLToMarkdownNesting(t *testing.T) {
	lines := strings.Repeat("x<br>", 2000)
	for name, level := range map[string][2]string{
		"blockquotes":   {"<blockquote>", "</blockquote>"},
		"bullet lists":  {"<ul><li>", "</li></ul>"},
		"ordered lists": {"<ol><li>", "</li></ol>"},
		"mixed":         {"<blockquote><ul><li>", "</li></ul></blockquote>"},
	} {
		t.Run("deep "+name+" render linearly", func(t *testing.T) {
			in := strings.Repeat(level[0], 150) + lines + strings.Repeat(level[1], 150)

			got, err := HTMLToMarkdown(in)
			require.NoError(t, err)
			assert.Contains(t, got, "x")
			assert.Less(t, len(got), 10*len(in))
		})
	}

	t.Run("shallow nesting keeps its structure", func(t *testing.T) {
		got, err := HTMLToMarkdown("<blockquote><blockquote><ul><li><ol><li>x</li></ol></li></ul></blockquote></blockquote>")
		require.NoError(t, err)
		assert.Equal(t, "> > - 1. x", got)
	})
}
