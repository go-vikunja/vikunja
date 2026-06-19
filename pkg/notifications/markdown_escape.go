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

import "strings"

// markdownSpecialChars is the full CommonMark §2.4 backslash-escapable set.
// '<' is included to neutralize autolinks (`<https://evil.com>`), which
// bluemonday's UGC policy would otherwise render as clickable <a> tags.
// '\' is handled separately first so inserted backslashes are not re-escaped.
const markdownSpecialChars = "`*_{}[]()<>#+-.!|~"

// EscapeMarkdown escapes every CommonMark-special character in s. Fixes
// GHSA-45q4-x4r9-8fqj (Markdown injection in notification emails).
func EscapeMarkdown(s string) string {
	// Backslash first so inserted backslashes are not double-escaped.
	s = strings.ReplaceAll(s, `\`, `\\`)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 128 && strings.ContainsRune(markdownSpecialChars, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
