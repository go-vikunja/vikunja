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

package csv

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func labelTitles(task *models.TaskWithComments) []string {
	titles := []string{}
	for _, l := range task.Labels {
		titles = append(titles, l.Title)
	}
	return titles
}

func labelsCell(count int) string {
	labels := make([]string, count)
	for i := range labels {
		labels[i] = "l" + strconv.Itoa(i)
	}
	return `"` + strings.Join(labels, ",") + `"`
}

func TestCSVLabels(t *testing.T) {
	importConfig := &ImportConfig{
		Delimiter: ",",
		SkipRows:  1,
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrLabels},
			{ColumnIndex: 2, Attribute: AttrLabels},
		},
	}
	convert := func(t *testing.T, content string) ([]*models.ProjectWithTasksAndBuckets, error) {
		t.Helper()
		return projectsFromCSV(bytes.NewReader([]byte(content)), int64(len(content)), importConfig)
	}

	t.Run("repeated labels of a row are imported once", func(t *testing.T) {
		projects, err := convert(t, "Title,Labels,More\nskipped,,\nTask,\"a, b, a\",\"b,c\"\n")
		require.NoError(t, err)
		require.Len(t, projects[0].Tasks, 1)
		assert.Equal(t, []string{"a", "b", "c"}, labelTitles(projects[0].Tasks[0]))
	})

	t.Run("the label limit per task is allowed", func(t *testing.T) {
		_, err := convert(t, "Title,Labels,More\nskipped,,\nTask,"+labelsCell(MaxLabelsPerTask)+",l0\n")
		require.NoError(t, err)
	})

	tooMany := "Title,Labels,More\nskipped,,\nTask,a,\nTask," + labelsCell(MaxLabelsPerTask) + ",extra\n"
	assertTooManyErr := func(t *testing.T, err error) {
		t.Helper()
		var labelsErr *migration.ErrCSVTooManyLabels
		require.ErrorAs(t, err, &labelsErr)
		assert.Equal(t, 4, labelsErr.Row, "row in the file, counting the header")
		assert.Equal(t, MaxLabelsPerTask, labelsErr.MaxLabels)
	}

	t.Run("more labels than the limit are rejected by the import", func(t *testing.T) {
		_, err := convert(t, tooMany)
		assertTooManyErr(t, err)
	})

	t.Run("more labels than the limit are rejected by the preview", func(t *testing.T) {
		_, err := PreviewImport(bytes.NewReader([]byte(tooMany)), int64(len(tooMany)), importConfig)
		assertTooManyErr(t, err)
	})
}

func TestCSVItemLimitDuringConversion(t *testing.T) {
	config.MigrationMaxItems.Set(10)
	t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })

	content := buildCSV(20)
	importConfig := &ImportConfig{
		Delimiter: ",",
		Mapping:   []ColumnMapping{{ColumnIndex: 0, Attribute: AttrTitle}},
	}
	_, err := projectsFromCSV(bytes.NewReader([]byte(content)), int64(len(content)), importConfig)
	var limitErr *migration.ErrImportItemLimitExceeded
	require.ErrorAs(t, err, &limitErr)
}
