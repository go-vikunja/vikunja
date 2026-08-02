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

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// The NewError override logs server errors; initialise a logger so that
	// path doesn't nil-panic in the bare test binary.
	log.InitLogger()
	os.Exit(m.Run())
}

// TestNewError_StripsServerErrorDetail guards against leaking internal error
// detail (raw DB/driver messages carrying hosts, ports, credentials) in v2 5xx
// responses. The strip lives in NewError because the huma.Error5xx* helpers
// call it directly, bypassing NewErrorWithContext, and huma writes an
// already-built StatusError as-is. Mirrors v1's generic-500 behaviour.
func TestNewError_StripsServerErrorDetail(t *testing.T) {
	secret := errors.New(`dial tcp 127.0.0.1:6390: connect: connection refused`)

	t.Run("Error500InternalServerError drops the wrapped detail", func(t *testing.T) {
		se := huma.Error500InternalServerError("Internal server error", secret)
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		assert.Empty(t, vm.Errors, "server errors must not expose internal detail")
		assert.Equal(t, "Internal server error", vm.Detail)

		body, err := json.Marshal(se)
		require.NoError(t, err)
		assert.NotContains(t, string(body), "6390", "serialized body must not carry the cause")
	})

	t.Run("Error503ServiceUnavailable drops the wrapped detail", func(t *testing.T) {
		se := huma.Error503ServiceUnavailable("service unavailable", secret)
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		assert.Empty(t, vm.Errors)
	})

	t.Run("NewErrorWithContext drops the wrapped detail", func(t *testing.T) {
		// Huma's handler-error path funnels raw errors through here at 500.
		se := huma.NewErrorWithContext(nil, 500, "unexpected error occurred", secret)
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		assert.Empty(t, vm.Errors, "server errors must not expose internal detail")
		assert.Equal(t, "unexpected error occurred", vm.Detail)
	})

	t.Run("4xx keeps the detail", func(t *testing.T) {
		se := huma.NewErrorWithContext(nil, 422, "validation failed", secret)
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		require.Len(t, vm.Errors, 1, "client errors keep their detail")
		assert.Equal(t, secret.Error(), vm.Errors[0].Message)
	})

	t.Run("4xx keeps ErrorDetailer locations", func(t *testing.T) {
		se := huma.Error422UnprocessableEntity("validation failed", &huma.ErrorDetail{Location: "body.title", Message: "cannot be empty"})
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		require.Len(t, vm.Errors, 1)
		assert.Equal(t, "body.title", vm.Errors[0].Location)
		assert.Equal(t, "cannot be empty", vm.Errors[0].Message)
	})

	t.Run("domain error code survives", func(t *testing.T) {
		// translateDomainError sets Code/I18nParams on the model NewError builds.
		se := translateDomainError(models.ErrLabelDoesNotExist{LabelID: 42})
		vm, ok := se.(*vikunjaErrorModel)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, vm.Status)
		assert.NotZero(t, vm.Code, "the Vikunja numeric error code must stay on the body")
	})
}

// The schema huma derives at registration time comes from NewError(0, ""), so
// the 5xx strip must not alter it.
func TestNewError_SchemaProbeUnaffected(t *testing.T) {
	se := huma.NewError(0, "")
	vm, ok := se.(*vikunjaErrorModel)
	require.True(t, ok)
	assert.Empty(t, vm.Errors)
	assert.Equal(t, 0, vm.Status)
}
