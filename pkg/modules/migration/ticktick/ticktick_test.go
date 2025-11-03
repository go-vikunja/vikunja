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

package ticktick

import (
	"bytes"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"
	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertTicktickTasksToVikunja(t *testing.T) {
	t1, err := time.Parse(time.RFC3339Nano, "2022-11-18T03:00:00.4770000Z")
	require.NoError(t, err)
	time1 := tickTickTime{Time: t1}
	t2, err := time.Parse(time.RFC3339Nano, "2022-12-18T03:00:00.4770000Z")
	require.NoError(t, err)
	time2 := tickTickTime{Time: t2}
	t3, err := time.Parse(time.RFC3339Nano, "2022-12-10T03:00:00.4770000Z")
	require.NoError(t, err)
	time3 := tickTickTime{Time: t3}
	duration, err := time.ParseDuration("24h")
	require.NoError(t, err)

	tickTickTasks := []*tickTickTask{
		{
			TaskID:      1,
			ParentID:    0,
			ProjectName: "Project 1",
			Title:       "Test task 1",
			Tags:        []string{"label1", "label2"},
			Content:     "Lorem Ipsum Dolor sit amet",
			StartDate:   time1,
			DueDate:     time2,
			Reminder:    duration,
			Repeat:      "FREQ=WEEKLY;INTERVAL=1;UNTIL=20190117T210000Z",
			Status:      "0",
			Order:       -1099511627776,
		},
		{
			TaskID:        2,
			ParentID:      1,
			ProjectName:   "Project 1",
			Title:         "Test task 2",
			Status:        "1",
			CompletedTime: time3,
			Order:         -1099511626,
		},
		{
			TaskID:      3,
			ParentID:    0,
			ProjectName: "Project 1",
			Title:       "Test task 3",
			Tags:        []string{"label1", "label2", "other label"},
			StartDate:   time1,
			DueDate:     time2,
			Reminder:    duration,
			Status:      "0",
			Order:       -109951627776,
		},
		{
			TaskID:      4,
			ParentID:    0,
			ProjectName: "Project 2",
			Title:       "Test task 4",
			Status:      "0",
			Order:       -109951627777,
		},
	}

	vikunjaTasks := convertTickTickToVikunja(tickTickTasks)

	assert.Len(t, vikunjaTasks, 3)

	assert.Equal(t, vikunjaTasks[1].ParentProjectID, vikunjaTasks[0].ID)
	assert.Equal(t, vikunjaTasks[2].ParentProjectID, vikunjaTasks[0].ID)

	assert.Len(t, vikunjaTasks[1].Tasks, 3)
	assert.Equal(t, vikunjaTasks[1].Title, tickTickTasks[0].ProjectName)

	assert.Equal(t, vikunjaTasks[1].Tasks[0].Title, tickTickTasks[0].Title)
	assert.Equal(t, vikunjaTasks[1].Tasks[0].Description, tickTickTasks[0].Content)
	assert.Equal(t, vikunjaTasks[1].Tasks[0].StartDate, tickTickTasks[0].StartDate.Time)
	assert.Equal(t, vikunjaTasks[1].Tasks[0].EndDate, tickTickTasks[0].DueDate.Time)
	assert.Equal(t, vikunjaTasks[1].Tasks[0].DueDate, tickTickTasks[0].DueDate.Time)
	assert.Equal(t, []*models.Label{
		{Title: "label1"},
		{Title: "label2"},
	}, vikunjaTasks[1].Tasks[0].Labels)
	assert.Equal(t, vikunjaTasks[1].Tasks[0].Reminders[0].RelativeTo, models.ReminderRelation("due_date"))
	assert.Equal(t, vikunjaTasks[1].Tasks[0].Reminders[0].RelativePeriod, int64(-24*3600))
	assert.Equal(t, vikunjaTasks[1].Tasks[0].Position, tickTickTasks[0].Order)
	assert.False(t, vikunjaTasks[1].Tasks[0].Done)

	assert.Equal(t, vikunjaTasks[1].Tasks[1].Title, tickTickTasks[1].Title)
	assert.Equal(t, vikunjaTasks[1].Tasks[1].Position, tickTickTasks[1].Order)
	assert.True(t, vikunjaTasks[1].Tasks[1].Done)
	assert.Equal(t, vikunjaTasks[1].Tasks[1].DoneAt, tickTickTasks[1].CompletedTime.Time)
	assert.Equal(t, models.RelatedTaskMap{
		models.RelationKindParenttask: []*models.Task{
			{
				ID: tickTickTasks[1].ParentID,
			},
		},
	}, vikunjaTasks[1].Tasks[1].RelatedTasks)

	assert.Equal(t, vikunjaTasks[1].Tasks[2].Title, tickTickTasks[2].Title)
	assert.Equal(t, vikunjaTasks[1].Tasks[2].Description, tickTickTasks[2].Content)
	assert.Equal(t, vikunjaTasks[1].Tasks[2].StartDate, tickTickTasks[2].StartDate.Time)
	assert.Equal(t, vikunjaTasks[1].Tasks[2].EndDate, tickTickTasks[2].DueDate.Time)
	assert.Equal(t, vikunjaTasks[1].Tasks[2].DueDate, tickTickTasks[2].DueDate.Time)
	assert.Equal(t, []*models.Label{
		{Title: "label1"},
		{Title: "label2"},
		{Title: "other label"},
	}, vikunjaTasks[1].Tasks[2].Labels)
	assert.Equal(t, vikunjaTasks[1].Tasks[2].Reminders[0].RelativeTo, models.ReminderRelation("due_date"))
	assert.Equal(t, vikunjaTasks[1].Tasks[2].Reminders[0].RelativePeriod, int64(-24*3600))
	assert.Equal(t, vikunjaTasks[1].Tasks[2].Position, tickTickTasks[2].Order)
	assert.False(t, vikunjaTasks[1].Tasks[2].Done)

	assert.Len(t, vikunjaTasks[2].Tasks, 1)
	assert.Equal(t, vikunjaTasks[2].Title, tickTickTasks[3].ProjectName)

	assert.Equal(t, vikunjaTasks[2].Tasks[0].Title, tickTickTasks[3].Title)
	assert.Equal(t, vikunjaTasks[2].Tasks[0].Position, tickTickTasks[3].Order)
}

func TestLinesToSkipBeforeHeader(t *testing.T) {
	csvContent := "Date: 2024-01-01+0000\nVersion: 7.1\n" +
		"\"Folder Name\",\"List Name\",\"Title\",\"Kind\",\"Tags\",\"Content\",\"Is Check list\",\"Start Date\",\"Due Date\",\"Reminder\",\"Repeat\",\"Priority\",\"Status\",\"Created Time\",\"Completed Time\",\"Order\",\"Timezone\",\"Is All Day\",\"Is Floating\",\"Column Name\",\"Column Order\",\"View Mode\",\"taskId\",\"parentId\"\n" +
		",\"list\",\"task1\",\"TEXT\",\"\",\"\",\"N\",\"\",\"\",\"\",\"\",\"0\",\"0\",\"2022-10-09T15:09:48+0000\",\"\",\"-1099511627776\",\"\",\"true\",\"false\",,,\"list\",\"1\",\"\"\n"

	r := bytes.NewReader([]byte(csvContent))
	lines, err := linesToSkipBeforeHeader(r, int64(len(csvContent)))
	require.NoError(t, err)
	assert.Equal(t, 2, lines)

	r2 := bytes.NewReader([]byte(csvContent))
	dec := newLineSkipDecoder(r2, lines)
	tasks := []*tickTickTask{}
	err = gocsv.UnmarshalDecoder(dec, &tasks)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "task1", tasks[0].Title)
}
