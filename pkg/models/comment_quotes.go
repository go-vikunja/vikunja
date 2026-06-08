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

// extractQuotedCommentIDs parses HTML and returns the set of positive integer
// comment ids referenced by <blockquote data-comment-id="…"> nodes anywhere in
// the document. Malformed values and zero/negative ids are silently skipped.
func extractQuotedCommentIDs(htmlText string) []int64 {
	if htmlText == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}

	ids := []int64{}
	seen := make(map[int64]bool)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "blockquote" {
			for _, attr := range n.Attr {
				if attr.Key != "data-comment-id" {
					continue
				}
				id, err := strconv.ParseInt(strings.TrimSpace(attr.Val), 10, 64)
				if err == nil && id > 0 && !seen[id] {
					ids = append(ids, id)
					seen[id] = true
				}
				break
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return ids
}

// findQuotedCommentAuthors resolves the authors of comments referenced via
// <blockquote data-comment-id="…"> within text, restricted to comments on the
// given task and excluding the doer themselves. Missing or wrong-task
// references are silently skipped. Link-share authors (negative ids) are
// also skipped since they cannot receive notifications.
func findQuotedCommentAuthors(s *xorm.Session, taskID, doerID int64, text string) (map[int64]*user.User, error) {
	ids := extractQuotedCommentIDs(text)
	if len(ids) == 0 {
		return nil, nil
	}

	rows := []*TaskComment{}
	err := s.
		In("id", ids).
		And("task_id = ?", taskID).
		Cols("id", "author_id", "task_id").
		Find(&rows)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	dedup := map[int64]bool{}
	wanted := []int64{}
	for _, r := range rows {
		id := r.AuthorID
		if id == doerID {
			continue
		}
		if id <= 0 {
			continue
		}
		if dedup[id] {
			continue
		}
		dedup[id] = true
		wanted = append(wanted, id)
	}
	if len(wanted) == 0 {
		return nil, nil
	}

	list := []*user.User{}
	err = s.In("id", wanted).Find(&list)
	if err != nil {
		return nil, err
	}

	users := make(map[int64]*user.User, len(list))
	for _, u := range list {
		users[u.ID] = u
	}
	return users, nil
}
