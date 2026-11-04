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

package planka

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func singleBoardData(bd *plankaBoardData) *plankaData {
	return &plankaData{
		Projects: []*plankaProjectData{{
			plankaProject: plankaProject{ID: "p1", Name: "Project"},
			Boards:        []*plankaBoardData{bd},
		}},
	}
}

func TestConvertDuplicateIDs(t *testing.T) {
	bd := &plankaBoardData{
		plankaBoard: plankaBoard{ID: "b1", Name: "Board"},
		boardIncluded: boardIncluded{
			Lists: []plankaList{
				{ID: "l1", Type: listTypeActive},
				{ID: "l1", Type: listTypeActive},
				{ID: "l1", Type: listTypeClosed},
			},
			Cards: []plankaCard{
				{ID: "c1", Name: "Card 1", ListID: "l1"},
				{ID: "c1", Name: "Card 1", ListID: "l1"},
				{ID: "c2", Name: "Card 2", ListID: "l1"},
			},
			TaskLists: []plankaTaskList{
				{ID: "tl1", Name: "Checklist", CardID: "c1"},
				{ID: "tl1", Name: "Checklist", CardID: "c1"},
			},
			Tasks: []plankaTask{{ID: "t1", Name: "Check item", TaskListID: "tl1"}},
		},
	}

	projects, err := convertPlankaToVikunja(singleBoardData(bd), nil)
	require.NoError(t, err)
	require.Len(t, projects, 2)

	project := projects[1]
	assert.Len(t, project.Buckets, 1)
	require.Len(t, project.Tasks, 2)
	assert.Equal(t, "Card 1", project.Tasks[0].Title)
	assert.Equal(t, "Card 2", project.Tasks[1].Title)
	assert.Equal(t, 1, strings.Count(project.Tasks[0].Description, "Check item"))
}

func TestConvertItemLimit(t *testing.T) {
	config.MigrationMaxItems.Set(10)
	t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })

	bd := &plankaBoardData{
		plankaBoard:   plankaBoard{ID: "b1", Name: "Board"},
		boardIncluded: boardIncluded{Lists: []plankaList{{ID: "l1", Type: listTypeActive}}},
	}
	for i := range 20 {
		id := strconv.Itoa(i)
		bd.Cards = append(bd.Cards, plankaCard{ID: "c" + id, Name: "Card " + id, ListID: "l1"})
		bd.Attachments = append(bd.Attachments, plankaAttachment{ID: "a" + id, CardID: "c" + id})
	}

	downloads := 0
	_, err := convertPlankaToVikunja(singleBoardData(bd), func(*plankaAttachment) (*bytes.Buffer, error) {
		downloads++
		return bytes.NewBufferString("x"), nil
	})
	var limitErr *migration.ErrImportItemLimitExceeded
	require.ErrorAs(t, err, &limitErr)
	assert.Less(t, downloads, 10, "conversion stops once the limit is reached")
}
