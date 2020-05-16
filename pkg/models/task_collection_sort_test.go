// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSortParamValidation(t *testing.T) {
	t.Run("Test valid order by", func(t *testing.T) {
		t.Run(orderAscending.String(), func(t *testing.T) {
			s := &sortParam{
				orderBy: orderAscending,
				sortBy:  "id",
			}
			err := s.validate()
			assert.NoError(t, err)
		})
		t.Run(orderDescending.String(), func(t *testing.T) {
			s := &sortParam{
				orderBy: orderDescending,
				sortBy:  "id",
			}
			err := s.validate()
			assert.NoError(t, err)
		})
	})
	t.Run("Test valid sort by", func(t *testing.T) {
		for _, test := range []string{
			taskPropertyID,
			taskPropertyTitle,
			taskPropertyDescription,
			taskPropertyDone,
			taskPropertyDoneAtUnix,
			taskPropertyDueDateUnix,
			taskPropertyCreatedByID,
			taskPropertyListID,
			taskPropertyRepeatAfter,
			taskPropertyPriority,
			taskPropertyStartDateUnix,
			taskPropertyEndDateUnix,
			taskPropertyHexColor,
			taskPropertyPercentDone,
			taskPropertyUID,
			taskPropertyCreated,
			taskPropertyUpdated,
			taskPropertyPosition,
		} {
			t.Run(test, func(t *testing.T) {
				s := &sortParam{
					orderBy: orderAscending,
					sortBy:  test,
				}
				err := s.validate()
				assert.NoError(t, err)
			})
		}
	})
	t.Run("Test invalid order by", func(t *testing.T) {
		s := &sortParam{
			orderBy: "somethingInvalid",
			sortBy:  "id",
		}
		err := s.validate()
		assert.Error(t, err)
		assert.True(t, IsErrInvalidSortOrder(err))
	})
	t.Run("Test invalid sort by", func(t *testing.T) {
		s := &sortParam{
			orderBy: orderAscending,
			sortBy:  "somethingInvalid",
		}
		err := s.validate()
		assert.Error(t, err)
		assert.True(t, IsErrInvalidTaskField(err))
	})
}
