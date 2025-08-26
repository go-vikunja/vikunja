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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// HandlerFunc defines the signature for business logic functions that use the wrapper
type HandlerFunc func(s *xorm.Session, u *user.User, c echo.Context) error

// WithDBAndUser wraps a handler function with common boilerplate:
// - Creates and manages database session
// - Retrieves current user from context
// - Handles basic error responses
// - Manages transactions for write operations (when needsTransaction is true)
func WithDBAndUser(handlerFunc HandlerFunc, needsTransaction bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create database session
		s := db.NewSession()
		defer s.Close()

		// Get current user
		u, err := user.GetCurrentUser(c)
		if err != nil {
			return HandleHTTPError(err)
		}

		// Execute the business logic
		err = handlerFunc(s, u, c)
		if err != nil {
			// If it's already an echo.HTTPError, return it directly
			if httpErr, ok := err.(*echo.HTTPError); ok {
				return httpErr
			}
			return HandleHTTPError(err)
		}

		// Commit transaction if needed
		if needsTransaction {
			if err := s.Commit(); err != nil {
				return HandleHTTPError(err)
			}
		}

		return nil
	}
}

// WithDB wraps a handler function with database session management only:
// - Creates and manages database session
// - Handles basic error responses
// - Manages transactions for write operations (when needsTransaction is true)
// This is useful for handlers that don't need user authentication
func WithDB(handlerFunc func(s *xorm.Session, c echo.Context) error, needsTransaction bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create database session
		s := db.NewSession()
		defer s.Close()

		// Execute the business logic
		err := handlerFunc(s, c)
		if err != nil {
			// If it's already an echo.HTTPError, return it directly
			if httpErr, ok := err.(*echo.HTTPError); ok {
				return httpErr
			}
			return HandleHTTPError(err)
		}

		// Commit transaction if needed
		if needsTransaction {
			if err := s.Commit(); err != nil {
				return HandleHTTPError(err)
			}
		}

		return nil
	}
}
