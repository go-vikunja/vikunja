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

package files

import (
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestBuildUploadResult(t *testing.T) {
	t.Run("maps a domain error to its numeric code", func(t *testing.T) {
		// ErrTaskAttachmentIsTooLarge is an HTTPErrorProcessor, so its Code must surface.
		r := BuildUploadResult(nil, []error{models.ErrTaskAttachmentIsTooLarge{Size: 99}})
		assert.Empty(t, r.Success)
		if assert.Len(t, r.Errors, 1) {
			assert.Equal(t, models.ErrCodeTaskAttachmentIsTooLarge, r.Errors[0].Code)
			assert.NotEmpty(t, r.Errors[0].Message)
		}
	})

	t.Run("plain error has no code, just the message", func(t *testing.T) {
		r := BuildUploadResult(nil, []error{errors.New("boom")})
		if assert.Len(t, r.Errors, 1) {
			assert.Zero(t, r.Errors[0].Code)
			assert.Equal(t, "boom", r.Errors[0].Message)
		}
	})

	t.Run("preserves success and failure order", func(t *testing.T) {
		success := []*models.TaskAttachment{{ID: 1}, {ID: 2}}
		r := BuildUploadResult(success, []error{errors.New("first"), errors.New("second")})
		assert.Equal(t, success, r.Success)
		if assert.Len(t, r.Errors, 2) {
			assert.Equal(t, "first", r.Errors[0].Message)
			assert.Equal(t, "second", r.Errors[1].Message)
		}
	})
}
