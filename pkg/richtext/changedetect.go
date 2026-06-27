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

import "strings"

// Changed reports whether inbound markdown differs semantically from stored
// rich-text HTML, so callers can skip rewriting unchanged fields (avoids CalDAV
// read-modify-write churning the HTML and bumping Updated). Both sides are
// canonicalized to markdown before comparing: HTML→markdown isn't an identity, so
// an HTML-domain compare would always report "changed". Errs to true.
func Changed(storedHTML, incomingMarkdown string) bool {
	stored, err := HTMLToMarkdown(storedHTML)
	if err != nil {
		return true
	}

	incoming, err := canonicalMarkdown(incomingMarkdown)
	if err != nil {
		return true
	}

	return normalizeMarkdown(stored) != normalizeMarkdown(incoming)
}

// HTMLIsEmpty treats "", "<p></p>" and whitespace-only markup as empty.
func HTMLIsEmpty(htmlInput string) bool {
	md, err := HTMLToMarkdown(htmlInput)
	if err != nil {
		return false
	}
	return md == ""
}

// canonicalMarkdown round-trips markdown through HTML so it matches the shape
// HTMLToMarkdown yields from stored HTML. No session needed: a <mention-user> tag
// and an inbound "@username" both reduce to "@username".
func canonicalMarkdown(md string) (string, error) {
	h, err := MarkdownToHTML(md)
	if err != nil {
		return "", err
	}
	return HTMLToMarkdown(h)
}

func normalizeMarkdown(md string) string {
	md = strings.ReplaceAll(md, "\r\n", "\n")
	md = strings.ReplaceAll(md, "\r", "\n")
	return strings.TrimSpace(md)
}
