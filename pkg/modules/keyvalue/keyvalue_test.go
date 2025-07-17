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

package keyvalue

import (
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/modules/keyvalue/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRememberReturnsExisting(t *testing.T) {
	store = memory.NewStorage()
	err := Put("foo", "bar")
	require.NoError(t, err)

	called := false
	val, err := Remember("foo", func() (interface{}, error) {
		called = true
		return "baz", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.False(t, called)
}

func TestRememberComputesAndStores(t *testing.T) {
	store = memory.NewStorage()

	called := 0
	val, err := Remember("foo", func() (interface{}, error) {
		called++
		return "bar", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.Equal(t, 1, called)

	v, exists, err := Get("foo")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "bar", v)

	val, err = Remember("foo", func() (interface{}, error) {
		called++
		return "baz", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.Equal(t, 1, called)
}

func TestRememberErrorDoesNotStore(t *testing.T) {
	store = memory.NewStorage()

	_, err := Remember("foo", func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	require.Error(t, err)
	_, exists, err2 := Get("foo")
	require.NoError(t, err2)
	assert.False(t, exists)
}
