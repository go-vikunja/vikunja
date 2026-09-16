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

package errorreport

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/web"

	"github.com/getsentry/sentry-go"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

type errProjectDoesNotExist struct {
	ID int64
}

func (e *errProjectDoesNotExist) Error() string {
	return fmt.Sprintf("project does not exist [ID: %d]", e.ID)
}

func (e *errProjectDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: 3001, Message: "This project does not exist."}
}

type errTaskDoesNotExist struct{}

func (e *errTaskDoesNotExist) Error() string {
	return "task does not exist"
}

func (e *errTaskDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: 4001, Message: "This task does not exist."}
}

func TestFingerprintDomainError(t *testing.T) {
	t.Run("wrapped domain error groups by its code", func(t *testing.T) {
		err := fmt.Errorf("could not read project: %w", &errProjectDoesNotExist{ID: 42})
		assert.Equal(t, []string{"vikunja", "3001"}, Fingerprint(err))
	})
	t.Run("the wrapping does not change the fingerprint", func(t *testing.T) {
		inner := &errProjectDoesNotExist{ID: 42}
		wrapped := fmt.Errorf("a: %w", fmt.Errorf("b: %w", inner))
		assert.Equal(t, Fingerprint(inner), Fingerprint(wrapped))
	})
	t.Run("different domain errors do not group together", func(t *testing.T) {
		assert.NotEqual(t,
			Fingerprint(&errProjectDoesNotExist{ID: 42}),
			Fingerprint(&errTaskDoesNotExist{}),
		)
	})
	t.Run("the id is not part of the fingerprint", func(t *testing.T) {
		assert.Equal(t,
			Fingerprint(&errProjectDoesNotExist{ID: 1}),
			Fingerprint(&errProjectDoesNotExist{ID: 999}),
		)
	})
}

func TestFingerprintDriverError(t *testing.T) {
	t.Run("postgres deadlock groups by sqlstate", func(t *testing.T) {
		err := fmt.Errorf("could not update task: %w", &pq.Error{
			Code:    "40P01",
			Message: "deadlock detected",
		})
		assert.Equal(t, []string{"postgres", "40P01", "deadlock detected"}, Fingerprint(err))
	})
	t.Run("postgres errors with different sqlstates do not group together", func(t *testing.T) {
		deadlock := &pq.Error{Code: "40P01", Message: "deadlock detected"}
		duplicate := &pq.Error{Code: "23505", Message: "duplicate key value violates unique constraint"}
		assert.NotEqual(t, Fingerprint(deadlock), Fingerprint(duplicate))
	})
	t.Run("the same mysql error on different columns groups together", func(t *testing.T) {
		permission := fmt.Errorf("could not read tasks: %w", &mysql.MySQLError{
			Number:  1054,
			Message: "Unknown column 'tp.permission' in 'field list'",
		})
		right := fmt.Errorf("could not read tasks: %w", &mysql.MySQLError{
			Number:  1054,
			Message: "Unknown column 'lu.right' in 'where clause'",
		})
		assert.Equal(t, []string{"mysql", "1054", "unknown column ? in ?"}, Fingerprint(permission))
		assert.Equal(t, Fingerprint(permission), Fingerprint(right))
	})
	t.Run("different mysql errors do not group together", func(t *testing.T) {
		unknownColumn := &mysql.MySQLError{Number: 1054, Message: "Unknown column 'tp.permission' in 'field list'"}
		deadlock := &mysql.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock"}
		assert.NotEqual(t, Fingerprint(unknownColumn), Fingerprint(deadlock))
	})
	t.Run("sqlite groups by code", func(t *testing.T) {
		err := fmt.Errorf("could not save task: %w", sqlite3.Error{
			Code:         sqlite3.ErrConstraint,
			ExtendedCode: sqlite3.ErrConstraintUnique,
		})
		fingerprint := Fingerprint(err)
		assert.Equal(t, "sqlite", fingerprint[0])
		assert.Equal(t, fmt.Sprintf("%d", sqlite3.ErrConstraint), fingerprint[1])
	})
}

func TestFingerprintGenericError(t *testing.T) {
	t.Run("ids and paths are normalised away", func(t *testing.T) {
		first := fmt.Errorf("wrapped: %w", errors.New("could not open file /var/lib/vikunja/files/12345 for user 42"))
		second := errors.New("could not open file /srv/data/files/98 for user 7")
		assert.Equal(t, Fingerprint(first), Fingerprint(second))
		assert.Equal(t, []string{"*errors.errorString", "could not open file ? for user ?"}, Fingerprint(first))
	})
	t.Run("mail addresses and uuids are normalised away", func(t *testing.T) {
		assert.Equal(t,
			Fingerprint(errors.New("could not notify test@example.com about 0195a2f3-1b7c-7c3a-9b6e-2f8a1c4d5e6f")),
			Fingerprint(errors.New("could not notify someone@vikunja.io about 7b2c0d11-aaaa-4bbb-8ccc-1d2e3f4a5b6c")),
		)
	})
	t.Run("different messages do not group together", func(t *testing.T) {
		assert.NotEqual(t,
			Fingerprint(errors.New("could not connect to the database")),
			Fingerprint(errors.New("could not send mail")),
		)
	})
	t.Run("the fingerprint is bounded", func(t *testing.T) {
		long := ""
		for i := 0; i < 100; i++ {
			long += "some words in a very long error message "
		}
		assert.LessOrEqual(t, len(Fingerprint(errors.New(long))[1]), maxNormalizedLen)
	})
}

func TestFingerprintFallsBackToDefault(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, []string{DefaultFingerprint}, Fingerprint(nil))
	})
	t.Run("error without a usable message", func(t *testing.T) {
		assert.Equal(t, []string{DefaultFingerprint}, Fingerprint(errors.New("42")))
	})
}

func TestFingerprintHTTPAndPanic(t *testing.T) {
	t.Run("echo errors group by status", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", echo.NewHTTPError(http.StatusBadGateway, "upstream is down"))
		assert.Equal(t, []string{"echo", "502", "upstream is down"}, Fingerprint(err))
	})
	t.Run("a panic does not group with the same error returned normally", func(t *testing.T) {
		inner := &errProjectDoesNotExist{ID: 42}
		panicked := &middleware.PanicStackError{Err: inner, Stack: []byte("goroutine 1 [running]")}
		assert.Equal(t, []string{"panic", "vikunja", "3001"}, Fingerprint(panicked))
		assert.NotEqual(t, Fingerprint(inner), Fingerprint(panicked))
	})
}

func TestTypeName(t *testing.T) {
	assert.Empty(t, TypeName(nil))
	assert.Equal(t, "*errorreport.errProjectDoesNotExist", TypeName(fmt.Errorf("wrapped: %w", &errProjectDoesNotExist{})))
	assert.Equal(t, "*mysql.MySQLError", TypeName(fmt.Errorf("wrapped: %w", &mysql.MySQLError{Number: 1054})))
}

func TestApply(t *testing.T) {
	applied := func(t *testing.T, apply func(scope *sentry.Scope)) *sentry.Event {
		t.Helper()
		scope := sentry.NewScope()
		apply(scope)
		return scope.ApplyToEvent(sentry.NewEvent(), nil, nil)
	}

	t.Run("sets the derived fingerprint and the error type tag", func(t *testing.T) {
		err := fmt.Errorf("could not read project: %w", &errProjectDoesNotExist{ID: 42})
		event := applied(t, func(scope *sentry.Scope) { Apply(scope, err) })

		assert.Equal(t, []string{"vikunja", "3001"}, event.Fingerprint)
		assert.Equal(t, "*errorreport.errProjectDoesNotExist", event.Tags["error.type"])
	})
	t.Run("keeps an explicit fingerprint", func(t *testing.T) {
		err := errors.New("failed to handle message")
		event := applied(t, func(scope *sentry.Scope) {
			ApplyFingerprint(scope, err, "message_handle_failed", "task.created", Normalize("task 42 does not exist"))
		})

		assert.Equal(t, []string{"message_handle_failed", "task.created", "task ? does not exist"}, event.Fingerprint)
		assert.Equal(t, "*errors.errorString", event.Tags["error.type"])
	})
	t.Run("falls back to the sentry default without a fingerprint", func(t *testing.T) {
		event := applied(t, func(scope *sentry.Scope) { ApplyFingerprint(scope, errors.New("boom")) })
		assert.Equal(t, []string{DefaultFingerprint}, event.Fingerprint)
	})
}
