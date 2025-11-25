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
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/user"

	"golang.org/x/net/html"
	"xorm.io/xorm"
)

func FindMentionedUsersInText(s *xorm.Session, text string) (users map[int64]*user.User, err error) {
	userIDs := extractMentionedUserIDs(text)
	if len(userIDs) == 0 {
		return
	}

	return user.GetUsersByIDs(s, userIDs)
}

// extractMentionedUserIDs parses HTML content and extracts user IDs from mention spans.
// It looks for <span class="mention" data-id="123"> elements and returns the user IDs.
func extractMentionedUserIDs(htmlText string) []int64 {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}

	var userIDs []int64
	seen := make(map[int64]bool) // Deduplicate user IDs

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			isMention := false
			var dataID string

			// Check if this span has class="mention" and extract data-id
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "mention") {
					isMention = true
				}
				if attr.Key == "data-id" {
					dataID = attr.Val
				}
			}

			// If this is a mention span with a valid data-id, extract the user ID
			if isMention && dataID != "" {
				if userID, err := strconv.ParseInt(dataID, 10, 64); err == nil {
					if !seen[userID] {
						userIDs = append(userIDs, userID)
						seen[userID] = true
					}
				}
			}
		}

		// Traverse child nodes
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(doc)
	return userIDs
}
