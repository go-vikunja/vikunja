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

package dump

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertFieldValue(t *testing.T) {
	t.Run("Float field conversions", func(t *testing.T) {
		t.Run("should return float64 as-is", func(t *testing.T) {
			result, err := convertFieldValue("position", 123.45, true)
			require.NoError(t, err)
			assert.InEpsilon(t, 123.45, result, 0.0001)
		})

		t.Run("should convert int to float64", func(t *testing.T) {
			result, err := convertFieldValue("position", 42, true)
			require.NoError(t, err)
			assert.InEpsilon(t, 42.0, result, 0.0001)
		})

		t.Run("should decode base64 string and convert to float", func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString([]byte("123.45"))
			result, err := convertFieldValue("position", encoded, true)
			require.NoError(t, err)
			assert.InEpsilon(t, 123.45, result, 0.0001)
		})

		t.Run("should handle non-base64 string and convert to float", func(t *testing.T) {
			result, err := convertFieldValue("position", "67.89", true)
			require.NoError(t, err)
			assert.InEpsilon(t, 67.89, result, 0.0001)
		})

		t.Run("should return error for invalid float string", func(t *testing.T) {
			_, err := convertFieldValue("position", "not-a-number", true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse double value")
		})

		t.Run("should return error for unexpected type", func(t *testing.T) {
			_, err := convertFieldValue("position", []int{1, 2, 3}, true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unexpected type for float field")
		})
	})

	t.Run("JSON field conversions", func(t *testing.T) {
		t.Run("should decode base64 string", func(t *testing.T) {
			jsonData := `{"key": "value"}`
			encoded := base64.StdEncoding.EncodeToString([]byte(jsonData))
			result, err := convertFieldValue("permissions", encoded, false)
			require.NoError(t, err)
			assert.JSONEq(t, jsonData, result.(string))
		})

		t.Run("should handle non-base64 string", func(t *testing.T) {
			jsonData := `{"key": "value"}`
			result, err := convertFieldValue("permissions", jsonData, false)
			require.NoError(t, err)
			assert.JSONEq(t, jsonData, result.(string))
		})

		t.Run("should return nil for 'null' string", func(t *testing.T) {
			result, err := convertFieldValue("bucket_configuration", "null", false)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("should return nil for 'NULL' string", func(t *testing.T) {
			result, err := convertFieldValue("bucket_configuration", "NULL", false)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("should return nil for 'Null' string", func(t *testing.T) {
			result, err := convertFieldValue("bucket_configuration", "Null", false)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("should return error for non-string type", func(t *testing.T) {
			_, err := convertFieldValue("permissions", 123, false)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected string for JSON field")
		})
	})

	t.Run("Base64 error handling", func(t *testing.T) {
		t.Run("should handle base64 decode error for float field", func(t *testing.T) {
			invalidBase64 := "invalid base64 with spaces and special chars!!!"
			result, err := convertFieldValue("position", invalidBase64, true)
			// Should not error on decode, but should error on parse float
			require.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse double value")
			assert.Nil(t, result)
		})

		t.Run("should handle base64 decode error for JSON field", func(t *testing.T) {
			// For JSON fields, CorruptInputError just returns the raw string
			invalidBase64 := "invalid base64 with spaces and special chars!!!"
			result, err := convertFieldValue("permissions", invalidBase64, false)
			require.NoError(t, err)
			assert.Equal(t, invalidBase64, result)
		})
	})

	t.Run("Edge cases", func(t *testing.T) {
		t.Run("should handle empty string for JSON field", func(t *testing.T) {
			result, err := convertFieldValue("permissions", "", false)
			require.NoError(t, err)
			assert.Empty(t, result)
		})

		t.Run("should handle empty string for float field", func(t *testing.T) {
			_, err := convertFieldValue("position", "", true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse double value")
		})

		t.Run("should handle zero float value", func(t *testing.T) {
			result, err := convertFieldValue("position", 0.0, true)
			require.NoError(t, err)
			assert.InDelta(t, 0.0, result, 0.0001)
		})

		t.Run("should handle negative float value", func(t *testing.T) {
			result, err := convertFieldValue("position", -123.45, true)
			require.NoError(t, err)
			assert.InEpsilon(t, -123.45, result, 0.0001)
		})
	})
}
