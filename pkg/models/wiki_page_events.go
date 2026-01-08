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
	"code.vikunja.io/api/pkg/user"
)

// WikiPageCreatedEvent represents an event where a wiki page was created
type WikiPageCreatedEvent struct {
	WikiPage *WikiPage
	Doer     *user.User
}

// Name returns the name of the event
func (w *WikiPageCreatedEvent) Name() string {
	return "wiki_page.created"
}

// WikiPageUpdatedEvent represents an event where a wiki page was updated
type WikiPageUpdatedEvent struct {
	WikiPage *WikiPage
	Doer     *user.User
}

// Name returns the name of the event
func (w *WikiPageUpdatedEvent) Name() string {
	return "wiki_page.updated"
}

// WikiPageDeletedEvent represents an event where a wiki page was deleted
type WikiPageDeletedEvent struct {
	WikiPage *WikiPage
	Doer     *user.User
}

// Name returns the name of the event
func (w *WikiPageDeletedEvent) Name() string {
	return "wiki_page.deleted"
}
