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

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/modules/migration/planka"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
)

// serverSideDomainError stands in for a domain error which maps to a 5xx: those are ours to fix.
type serverSideDomainError struct{}

func (e *serverSideDomainError) Error() string { return "something broke on our end" }

func (e *serverSideDomainError) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusInternalServerError,
		Code:     9999,
		Message:  "something broke on our end",
	}
}

func TestShouldReportMigrationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "upstream 404",
			err:  migration.NewErrUpstreamRequestFailed("microsoft graph", 404, `{"error":{"code":"itemNotFound"}}`),
			want: false,
		},
		{
			name: "upstream 401",
			err:  migration.NewErrUpstreamRequestFailed("todoist oauth", 401, "unauthorized"),
			want: false,
		},
		{
			name: "wrapped upstream 404",
			err:  fmt.Errorf("could not get data: %w", migration.NewErrUpstreamRequestFailed("microsoft graph", 404, "")),
			want: false,
		},
		{
			name: "upstream 500",
			err:  migration.NewErrUpstreamRequestFailed("planka", 500, "boom"),
			want: true,
		},
		{
			name: "wrapped upstream 503",
			err:  fmt.Errorf("could not get data: %w", migration.NewErrUpstreamRequestFailed("planka", 503, "")),
			want: true,
		},
		{
			name: "planka invalid credentials",
			err:  &planka.ErrInvalidCredentials{},
			want: false,
		},
		{
			name: "planka no api at url",
			err:  &planka.ErrNoPlankaAtURL{Reason: "404"},
			want: false,
		},
		{
			name: "wrapped 4xx domain error",
			err:  fmt.Errorf("could not migrate: %w", &migration.ErrNotAZipFile{}),
			want: false,
		},
		{
			name: "5xx domain error",
			err:  &serverSideDomainError{},
			want: true,
		},
		{
			name: "arbitrary error",
			err:  errors.New("something went wrong"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldReportMigrationError(tt.err))
		})
	}
}
