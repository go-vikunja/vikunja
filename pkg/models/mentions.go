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

package models

import (
	"strings"

	"code.vikunja.io/api/pkg/user"

	"golang.org/x/net/html"
	"xorm.io/xorm"
)

func FindMentionedUsersInText(s *xorm.Session, text string) (users map[int64]*user.User, err error) {
	usernames := extractMentionedUsernames(text)
	if len(usernames) == 0 {
		return
	}

	return user.GetUsersByUsername(s, usernames, true)
}

// extractMentionedUsernames parses HTML content and extracts usernames from mention spans.
// It looks for <mention-user data-id="username"> elements and returns the usernames.
func extractMentionedUsernames(htmlText string) []string {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}

	usernames := []string{}
	seen := make(map[string]bool) // Deduplicate usernames

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "mention-user" {
			var dataID string

			// Extract data-id attribute
			for _, attr := range n.Attr {
				if attr.Key == "data-id" {
					dataID = attr.Val
				}
			}

			if dataID != "" {
				if !seen[dataID] {
					usernames = append(usernames, dataID)
					seen[dataID] = true
				}
			}
		}

		// Traverse child nodes
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(doc)
	return usernames
}
