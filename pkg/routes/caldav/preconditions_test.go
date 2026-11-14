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

package caldav

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreconditions(t *testing.T) {
	const etag = `"40-1700000000"`

	tests := []struct {
		name                string
		header              http.Header
		create, update, del bool
	}{
		{name: "no headers", header: http.Header{}, create: true, update: true, del: true},
		{name: "If-Match matching", header: http.Header{"If-Match": {etag}}, update: true, del: true},
		{name: "If-Match star", header: http.Header{"If-Match": {"*"}}, update: true, del: true},
		{name: "If-Match stale", header: http.Header{"If-Match": {`"stale"`}}},
		{name: "If-Match unquoted", header: http.Header{"If-Match": {"40-1700000000"}}},
		{name: "If-Match list containing the etag", header: http.Header{"If-Match": {`"stale", ` + etag}}},
		{name: "If-Match only first value counts", header: http.Header{"If-Match": {`"stale"`, etag}}},
		{name: "If-Match first value matches", header: http.Header{"If-Match": {etag, `"stale"`}}, update: true, del: true},
		{name: "If-None-Match star", header: http.Header{"If-None-Match": {"*"}}, create: true, del: true},
		{name: "If-None-Match star twice", header: http.Header{"If-None-Match": {"*", "*"}}, create: true, update: true, del: true},
		{name: "If-None-Match matching etag", header: http.Header{"If-None-Match": {etag}}, create: true, update: true, del: true},
		{name: "If-Match matching and If-None-Match star", header: http.Header{"If-Match": {etag}, "If-None-Match": {"*"}}, del: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := preconditionsOf(tc.header)
			assert.Equal(t, tc.create, p.allowCreate(), "create")
			assert.Equal(t, tc.update, p.allowUpdate(etag), "update")
			assert.Equal(t, tc.del, p.allowDelete(etag), "delete")
		})
	}
}
