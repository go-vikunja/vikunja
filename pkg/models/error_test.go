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

package models

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var i18nPlaceholderPattern = regexp.MustCompile(`\{(\w+)}`)

func loadI18nErrorMessages(t *testing.T) map[string]string {
	t.Helper()

	raw, err := os.ReadFile("../../frontend/src/i18n/lang/en.json")
	require.NoError(t, err)

	var lang struct {
		Error map[string]string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(raw, &lang))

	return lang.Error
}

func placeholderNames(message string) map[string]bool {
	names := map[string]bool{}
	for _, match := range i18nPlaceholderPattern.FindAllStringSubmatch(message, -1) {
		names[match[1]] = true
	}
	return names
}

// TestI18nParamsContract asserts that each error's I18nParams keys are exactly the
// {placeholder} names used in the corresponding error message in
// frontend/src/i18n/lang/en.json, so a placeholder rename on either side fails the test.
func TestI18nParamsContract(t *testing.T) {
	i18nMessages := loadI18nErrorMessages(t)

	tests := []struct {
		name         string
		err          web.HTTPErrorProcessor
		expectedCode int
	}{
		{
			name:         "ErrInvalidTimezone",
			err:          ErrInvalidTimezone{Name: "x"},
			expectedCode: ErrCodeInvalidTimezone,
		},
		{
			name:         "ErrInvalidAPITokenPermission",
			err:          &ErrInvalidAPITokenPermission{Group: "g", Permission: "p"},
			expectedCode: ErrCodeInvalidAPITokenPermission,
		},
		{
			name:         "user.ErrInvalidClaimData",
			err:          &user.ErrInvalidClaimData{Field: "f", Type: "t"},
			expectedCode: user.ErrCodeInvalidClaimData,
		},
		{
			name:         "user.ErrInvalidTimezone",
			err:          user.ErrInvalidTimezone{Name: "x"},
			expectedCode: user.ErrorCodeInvalidTimezone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := tt.err.HTTPError()
			assert.Equal(t, tt.expectedCode, httpErr.Code)

			message, ok := i18nMessages[strconv.Itoa(tt.expectedCode)]
			require.True(t, ok, "no en.json error message found for code %d", tt.expectedCode)

			expectedParams := placeholderNames(message)
			actualParams := map[string]bool{}
			for key := range httpErr.I18nParams {
				actualParams[key] = true
			}

			assert.Equal(t, expectedParams, actualParams)
		})
	}

	coveredCodes := map[string]bool{}
	for _, tt := range tests {
		coveredCodes[strconv.Itoa(tt.expectedCode)] = true
	}

	parametrisedCodes := map[string]bool{}
	for code, message := range i18nMessages {
		if len(placeholderNames(message)) > 0 {
			parametrisedCodes[code] = true
		}
	}

	assert.Equal(t, parametrisedCodes, coveredCodes, "every parametrised en.json error message must have a corresponding I18nParams test case")
}
