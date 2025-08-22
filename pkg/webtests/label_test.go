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

package webtests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getIDFromJSON(t *testing.T, body string) string {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	require.NoError(t, err)
	return fmt.Sprintf("%v", data["id"])
}

func TestLabelAPI(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	// Create a label
	payload := `{"title":"My Test Label","description":"A test label","hex_color":"ff00ff"}`
	rec, err := th.Request(t, "POST", "/api/v1/labels", strings.NewReader(payload))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"title":"My Test Label"`)
	assert.Contains(t, body, `"description":"A test label"`)
	assert.Contains(t, body, `"hex_color":"ff00ff"`)

	// Extract the id of the new label
	id := getIDFromJSON(t, body)

	// Get the label
	rec, err = th.Request(t, "GET", "/api/v1/labels/"+id, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"title":"My Test Label"`)

	// Get all labels
	rec, err = th.Request(t, "GET", "/api/v1/labels", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), `"title":"My Test Label"`))

	// Update the label
	updatePayload := `{"title":"My Updated Label"}`
	rec, err = th.Request(t, "PUT", "/api/v1/labels/"+id, strings.NewReader(updatePayload))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"title":"My Updated Label"`)

	// Delete the label
	rec, err = th.Request(t, "DELETE", "/api/v1/labels/"+id, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify the label is gone
	rec, err = th.Request(t, "GET", "/api/v1/labels/"+id, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
