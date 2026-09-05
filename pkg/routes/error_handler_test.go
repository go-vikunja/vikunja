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

package routes

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHTTPErrorHandler(t *testing.T) {
	forbidden := models.ErrUserDoesNotHaveAccessToProject{ProjectID: 27215, UserID: 1}
	const forbiddenBody = `{"code":7003,"message":"This user does not have access to the project."}`

	validation := models.InvalidFieldError([]string{"title"})
	const validationBody = `{"code":2002,"message":"Invalid Data","invalid_fields":["title"]}`

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "domain error",
			err:      forbidden,
			wantCode: http.StatusForbidden,
			wantBody: forbiddenBody,
		},
		{
			name:     "domain error wrapped once",
			err:      fmt.Errorf("could not insert data: %w", forbidden),
			wantCode: http.StatusForbidden,
			wantBody: forbiddenBody,
		},
		{
			name:     "domain error wrapped twice",
			err:      fmt.Errorf("could not migrate: %w", fmt.Errorf("could not insert data: %w", forbidden)),
			wantCode: http.StatusForbidden,
			wantBody: forbiddenBody,
		},
		{
			name:     "validation error",
			err:      validation,
			wantCode: http.StatusPreconditionFailed,
			wantBody: validationBody,
		},
		{
			name:     "validation error wrapped",
			err:      fmt.Errorf("could not create task: %w", validation),
			wantCode: http.StatusPreconditionFailed,
			wantBody: validationBody,
		},
		{
			name:     "unknown error",
			err:      errors.New("something went wrong"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"Internal Server Error"}`,
		},
		{
			name:     "echo http error wrapped",
			err:      fmt.Errorf("could not handle request: %w", echo.NewHTTPError(http.StatusForbidden, "Forbidden")),
			wantCode: http.StatusForbidden,
			wantBody: `{"message":"Forbidden"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.InitLogger()
			e := echo.New()
			e.HTTPErrorHandler = CreateHTTPErrorHandler(e, false)
			e.GET("/test", func(_ *echo.Context) error {
				return tt.err
			})

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

			assert.Equal(t, tt.wantCode, rec.Code)
			require.NotEmpty(t, rec.Body.String())
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
