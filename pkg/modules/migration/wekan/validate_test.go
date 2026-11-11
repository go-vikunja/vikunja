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

package wekan

import (
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateBoard(t *testing.T, board string) error {
	t.Helper()
	r := strings.NewReader(board)
	return (&Migrator{}).ValidateFile(r, int64(r.Len()))
}

func TestValidateFile(t *testing.T) {
	t.Run("accepts the fixture export", func(t *testing.T) {
		export, err := os.ReadFile("testdata_wekan_export.json")
		require.NoError(t, err)
		require.NoError(t, validateBoard(t, string(export)))
	})

	for kind, board := range map[string]string{
		"card":      `{"title":"b","cards":[{"_id":"dup"},{"_id":"dup"}]}`,
		"checklist": `{"title":"b","cards":[{"_id":"c1"}],"checklists":[{"_id":"dup","cardId":"c1"},{"_id":"dup","cardId":"c1"}]}`,
		"comment":   `{"title":"b","cards":[{"_id":"c1"}],"comments":[{"_id":"dup","cardId":"c1"},{"_id":"dup","cardId":"c1"}]}`,
	} {
		t.Run("rejects duplicate "+kind+" ids", func(t *testing.T) {
			err := validateBoard(t, board)
			var invalid *migration.ErrInvalidImportFile
			require.ErrorAs(t, err, &invalid)
			assert.Contains(t, err.Error(), "duplicate "+kind+" id")
		})
	}

	t.Run("rejects more items than migration.maxitems", func(t *testing.T) {
		config.MigrationMaxItems.Set(2)
		t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })

		err := validateBoard(t, `{"title":"b","lists":[{"_id":"l1"}],"cards":[{"_id":"c1"},{"_id":"c2"}]}`)
		var limitErr *migration.ErrImportItemLimitExceeded
		require.ErrorAs(t, err, &limitErr)
	})

	t.Run("counts card label references toward migration.maxitems", func(t *testing.T) {
		config.MigrationMaxItems.Set(3)
		t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })

		err := validateBoard(t, `{"title":"b","labels":[{"_id":"l1"}],"cards":[{"_id":"c1","labelIds":["l1","l1","l1"]}]}`)
		var limitErr *migration.ErrImportItemLimitExceeded
		require.ErrorAs(t, err, &limitErr)
	})

	t.Run("leaves an undecodable document to the import job", func(t *testing.T) {
		require.NoError(t, validateBoard(t, `{"title": `))
	})
}

func TestConvertWekanManyChecklistItems(t *testing.T) {
	items := make([]wekanChecklistItem, 0, 3)
	for _, title := range []string{"a", "b", "c"} {
		items = append(items, wekanChecklistItem{ChecklistID: "cl1", Title: title})
	}
	board := &wekanBoard{
		Cards:          []wekanCard{{ID: "c1", Title: "card", Description: "desc"}},
		Checklists:     []wekanChecklist{{ID: "cl1", CardID: "c1", Title: "list"}},
		ChecklistItems: items,
	}

	desc := convertWekanToVikunja(board)[0].Tasks[0].Description
	assert.Equal(t, 1, strings.Count(desc, "<h2> list</h2>"))
	assert.Equal(t, 3, strings.Count(desc, `data-type="taskItem"`))
	assert.True(t, strings.HasPrefix(desc, "<p>desc</p>"), desc)
}
