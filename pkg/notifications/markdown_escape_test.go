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

package notifications

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain ASCII", "Hello World", "Hello World"},
		{"backslash", `a\b`, `a\\b`},
		{"link open bracket", "a[b", `a\[b`},
		{"link close bracket", "a]b", `a\]b`},
		{"paren open", "a(b", `a\(b`},
		{"paren close", "a)b", `a\)b`},
		{"image bang", "a!b", `a\!b`},
		{"emphasis asterisk", "a*b", `a\*b`},
		{"emphasis underscore", "a_b", `a\_b`},
		{"code backtick", "a`b", "a\\`b"},
		{"heading hash", "a#b", `a\#b`},
		{"blockquote", "a>b", `a\>b`},
		{"list dash", "a-b", `a\-b`},
		{"list plus", "a+b", `a\+b`},
		{"list dot", "a.b", `a\.b`},
		{"pipe (tables)", "a|b", `a\|b`},
		{"tilde (strikethrough)", "a~b", `a\~b`},
		{"curly brace open", "a{b", `a\{b`},
		{"curly brace close", "a}b", `a\}b`},
		{"angle bracket open", "a<b", `a\<b`},
		{"advisory PoC: fake link", "test](https://evil.com) [Click to verify", `test\]\(https://evil\.com\) \[Click to verify`},
		{"advisory PoC: tracking pixel", "![](https://evil.com/track.png)", `\!\[\]\(https://evil\.com/track\.png\)`},
		{"commonmark autolink", "<https://evil.com>", `\<https://evil\.com\>`},
		{"raw html tag", `<a href="https://evil.com">x</a>`, `\<a href="https://evil\.com"\>x\</a\>`},
		{"raw img tag", `<img src=x onerror=alert(1)>`, `\<img src=x onerror=alert\(1\)\>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeMarkdown(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestEscapeMarkdown_RoundTripThroughGoldmark verifies that every escaped
// string, when fed into goldmark as the text portion of a Markdown link,
// renders as the original literal text and does NOT produce any additional
// <a> or <img> elements from injected Markdown syntax.
func TestEscapeMarkdown_RoundTripThroughGoldmark(t *testing.T) {
	payloads := []string{
		"test](https://evil.com) [Click to verify",
		"![](https://evil.com/track.png)",
		"plain title",
		`a\b`,
		"`code`",
		"*bold*",
		"<https://evil.com>",
		`<a href="https://evil.com">click</a>`,
		`<img src=x onerror=alert(1)>`,
	}
	for _, p := range payloads {
		t.Run(p, func(t *testing.T) {
			// Embed in a markdown link and a free paragraph.
			md := "* [" + EscapeMarkdown(p) + "](https://vikunja.io/safe)\n\n" + EscapeMarkdown(p)
			var buf bytes.Buffer
			require.NoError(t, goldmark.Convert([]byte(md), &buf))
			html := buf.String()
			// There must be exactly one <a href=, the safe one.
			assert.Equal(t, 1, bytes.Count([]byte(html), []byte("<a href=")),
				"goldmark output for %q must contain exactly one <a href=: %s", p, html)
			// There must be zero <img tags — payloads that try to inject images must be escaped.
			assert.NotContains(t, html, "<img ", "goldmark output for %q must not contain an <img> tag: %s", p, html)
			// The safe URL must still be present.
			assert.Contains(t, html, `href="https://vikunja.io/safe"`)
		})
	}
}
