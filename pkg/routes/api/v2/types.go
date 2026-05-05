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

package apiv2

// Paginated is the standard list-response envelope for every /api/v2 list operation.
type Paginated[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int64 `json:"total_pages"`
}

// NewPaginated builds a Paginated envelope. Nil items become an empty
// slice so the JSON response is [] rather than null.
func NewPaginated[T any](items []T, total int64, page, perPage int) Paginated[T] {
	if items == nil {
		items = []T{}
	}
	var totalPages int64
	if perPage > 0 {
		totalPages = (total + int64(perPage) - 1) / int64(perPage)
	}
	return Paginated[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// ListParams carries the standard (page, per_page, q) query shape for list operations.
type ListParams struct {
	Page    int    `query:"page"     default:"1"  minimum:"1"`
	PerPage int    `query:"per_page" default:"50" minimum:"1" maximum:"1000"`
	Q       string `query:"q"`
}

// singleBody is the create/update response envelope (no ETag).
type singleBody[T any] struct {
	Body *T
}

// singleReadBody is the read response envelope; carries ETag for If-None-Match.
type singleReadBody[T any] struct {
	ETag string `header:"ETag"`
	Body *T
}

// emptyBody marks delete / no-content operations.
type emptyBody struct{}
