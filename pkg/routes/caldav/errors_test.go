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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorStatus(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "failed precondition",
			err:    fmt.Errorf("report: %w", errUnsupportedReport),
			status: http.StatusForbidden,
		},
		{
			name:   "validation",
			err:    models.ValidationHTTPError{},
			status: http.StatusForbidden,
		},
		{
			name:   "unknown",
			err:    errors.New("database is gone"),
			status: http.StatusInternalServerError,
		},
		{
			name:   "forbidden without read access is not mapped",
			err:    models.ErrGenericForbidden{},
			status: http.StatusInternalServerError,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, errorStatus(tc.err))
		})
	}
}

func TestWriteError(t *testing.T) {
	run := func(method string, err error) (*httptest.ResponseRecorder, error) {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest(method, "/dav/projects/1", nil), rec)
		return rec, writeError(c, err)
	}

	t.Run("client error answers directly", func(t *testing.T) {
		rec, err := run(http.MethodGet, errNotFound)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "Not Found", rec.Body.String())
	})

	t.Run("HEAD has no body", func(t *testing.T) {
		rec, err := run(http.MethodHead, errForbidden)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("unknown error goes to the error handler without its message", func(t *testing.T) {
		cause := errors.New("database is gone")
		_, err := run(http.MethodGet, cause)
		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Equal(t, "internal server error", he.Message)
		assert.ErrorIs(t, err, cause)
	})
}
