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
	"errors"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/log"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// The NewErrorWithContext override logs server errors; initialise a
	// logger so that path doesn't nil-panic in the bare test binary.
	log.InitLogger()
	os.Exit(m.Run())
}

// TestNewErrorWithContext_StripsServerErrorDetail guards against leaking
// internal error detail (raw DB/driver messages, etc.) in v2 5xx responses.
// Huma's handler-error path funnels raw errors through NewErrorWithContext
// at status 500; the override must drop the detail there while keeping it
// for client (4xx) errors. Mirrors v1's generic-500 behaviour.
func TestNewErrorWithContext_StripsServerErrorDetail(t *testing.T) {
	secret := errors.New(`pq: relation "labels" does not exist`)

	t.Run("500 drops the wrapped detail", func(t *testing.T) {
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
}
