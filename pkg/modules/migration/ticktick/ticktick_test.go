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
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"strings"
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
			Status:      "2",
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
	assert.True(t, vikunjaTasks[2].Tasks[0].Done)
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
	dec, err := newLineSkipDecoder(r2, lines)
	require.NoError(t, err)
	tasks := []*tickTickTask{}
	err = gocsv.UnmarshalDecoder(dec, &tasks)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "task1", tasks[0].Title)
}

func TestLinesToSkipBeforeHeaderWithRealCSV(t *testing.T) {
	// This is the actual format from a real TickTick export with BOM and multi-line status
	csvContent := "\uFEFF\"Date: 2025-11-25+0000\"\n" +
		"\"Version: 7.1\"\n" +
		"\"Status: \n" +
		"0 Normal\n" +
		"1 Completed\n" +
		"2 Archived\"\n" +
		"\"Folder Name\",\"List Name\",\"Title\",\"Kind\",\"Tags\",\"Content\",\"Is Check list\",\"Start Date\",\"Due Date\",\"Reminder\",\"Repeat\",\"Priority\",\"Status\",\"Created Time\",\"Completed Time\",\"Order\",\"Timezone\",\"Is All Day\",\"Is Floating\",\"Column Name\",\"Column Order\",\"View Mode\",\"taskId\",\"parentId\"\n" +
		"\"dsx\",\"x\",\"this task repeats\",\"TEXT\",\"\",\"\",\"N\",\"\",\"\",\"\",\"\",\"0\",\"0\",\"2022-10-09T15:09:48+0000\",\"\",\"-1099511627776\",\"Europe/Berlin\",,\"false\",,,\"list\",\"2\",\"\"\n"

	t.Logf("CSV content length: %d", len(csvContent))
	t.Logf("CSV content first 100 chars: %q", csvContent[:100])

	r := bytes.NewReader([]byte(csvContent))
	lines, err := linesToSkipBeforeHeader(r, int64(len(csvContent)))
	require.NoError(t, err)
	t.Logf("Lines to skip: %d", lines)
	assert.Equal(t, 6, lines) // Should skip 6 lines to get to the header

	r2 := bytes.NewReader([]byte(csvContent))
	dec, err := newLineSkipDecoder(r2, lines)
	require.NoError(t, err)
	tasks := []*tickTickTask{}
	err = gocsv.UnmarshalDecoder(dec, &tasks)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "this task repeats", tasks[0].Title)
	assert.Equal(t, "dsx", tasks[0].FolderName)
	assert.Equal(t, "x", tasks[0].ProjectName)
}

func TestLinesToSkipBeforeHeaderWithCleanTestFile(t *testing.T) {
	// Test with the cleaned-up test CSV file
	file, err := os.Open("testdata_ticktick_export.csv")
	require.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	require.NoError(t, err)

	lines, err := linesToSkipBeforeHeader(file, stat.Size())
	require.NoError(t, err)
	t.Logf("Lines to skip in test file: %d", lines)
	assert.Equal(t, 6, lines) // Should skip 6 lines to get to the header

	// Reset file position
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)

	// Let's manually check what the header line looks like after skipping
	r := stripBOM(file)
	scanner := bufio.NewScanner(r)
	for i := 0; i <= lines; i++ {
		if !scanner.Scan() {
			break
		}
		if i == lines {
			t.Logf("Header line after skipping %d lines: %q", lines, scanner.Text())
		}
	}

	// Reset file position again
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)

	dec, err := newLineSkipDecoder(file, lines)
	require.NoError(t, err)
	tasks := []*tickTickTask{}
	err = gocsv.UnmarshalDecoder(dec, &tasks)
	require.NoError(t, err)
	require.Greater(t, len(tasks), 0)

	// Verify that the first task has actual data
	assert.Equal(t, "Work", tasks[0].FolderName)
	assert.Equal(t, "Project Alpha", tasks[0].ProjectName)
	assert.Equal(t, "Task with repeating schedule", tasks[0].Title)
}

func TestBOMStripping(t *testing.T) {
	// Test BOM stripping specifically
	csvWithBOM := "\uFEFF\"Folder Name\",\"List Name\",\"Title\"\n\"test\",\"list\",\"task\"\n"

	r := stripBOM(bytes.NewReader([]byte(csvWithBOM)))
	scanner := bufio.NewScanner(r)

	// Read first line (header)
	require.True(t, scanner.Scan())
	header := scanner.Text()
	t.Logf("Header after BOM stripping: %q", header)

	// Read second line (data)
	require.True(t, scanner.Scan())
	data := scanner.Text()
	t.Logf("Data line: %q", data)

	// Test CSV parsing
	r2 := stripBOM(bytes.NewReader([]byte(csvWithBOM)))
	reader := csv.NewReader(r2)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	t.Logf("CSV records: %+v", records)
}

func TestEmptyLabelHandling(t *testing.T) {
	t.Run("Normal tags", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "work, personal, urgent",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{"work", "personal", "urgent"}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Tags with extra spaces", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "work,  personal  , urgent",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{"work", "personal", "urgent"}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Empty tags mixed with valid ones", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "work, , urgent, ",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{"work", "urgent"}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Only whitespace tags", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    " ,  ,   ",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Empty string", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Single valid tag", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "important",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{"important"}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Single empty tag", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    " ",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})

	t.Run("Tags with leading/trailing spaces", func(t *testing.T) {
		task := &tickTickTask{
			Title:       "Test Task",
			ProjectName: "Test Project",
			TagsList:    "  work  , personal,   urgent   ",
		}
		task.Tags = strings.Split(task.TagsList, ", ")

		vikunjaTasks := convertTickTickToVikunja([]*tickTickTask{task})
		projectWithTasks := findProjectWithTasks(t, vikunjaTasks)
		vikunjaTask := projectWithTasks.Tasks[0]

		expectedTags := []string{"work", "personal", "urgent"}
		assertLabelsMatch(t, vikunjaTask, expectedTags)
	})
}

// Helper function to find the project that contains tasks
func findProjectWithTasks(t *testing.T, vikunjaTasks []*models.ProjectWithTasksAndBuckets) *models.ProjectWithTasksAndBuckets {
	t.Helper()

	// The function creates a parent project and child projects
	// We expect 2 projects: parent "Migrated from TickTick" and child "Test Project"
	require.Len(t, vikunjaTasks, 2)

	// Find the project with tasks (should be the child project)
	for _, project := range vikunjaTasks {
		if len(project.Tasks) > 0 {
			require.Len(t, project.Tasks, 1)
			return project
		}
	}

	t.Fatal("Should find a project with tasks")
	return nil
}

// Helper function to assert that labels match expected tags
func assertLabelsMatch(t *testing.T, vikunjaTask *models.TaskWithComments, expectedTags []string) {
	t.Helper()

	// Check that only non-empty labels were created
	assert.Len(t, vikunjaTask.Labels, len(expectedTags), "Number of labels should match expected")

	// Check that the label titles match expected tags
	actualTags := make([]string, len(vikunjaTask.Labels))
	for i, label := range vikunjaTask.Labels {
		actualTags[i] = label.Title
	}

	assert.ElementsMatch(t, expectedTags, actualTags, "Label titles should match expected tags")

	// Ensure no empty labels were created
	for _, label := range vikunjaTask.Labels {
		assert.NotEmpty(t, strings.TrimSpace(label.Title), "No label should be empty or whitespace-only")
	}
}

func TestMultilineDescriptions(t *testing.T) {
	// Test with a CSV fixture that contains actual multiline content in quoted fields
	file, err := os.Open("testdata_ticktick_multiline.csv")
	require.NoError(t, err, "Failed to open test fixture")
	defer file.Close()

	stat, err := file.Stat()
	require.NoError(t, err)

	lines, err := linesToSkipBeforeHeader(file, stat.Size())
	require.NoError(t, err)
	t.Logf("Lines to skip: %d", lines)
	assert.Equal(t, 6, lines, "Should skip 6 metadata lines")

	// Reset file position
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)

	dec, err := newLineSkipDecoder(file, lines)
	require.NoError(t, err)
	tasks := []*tickTickTask{}
	err = gocsv.UnmarshalDecoder(dec, &tasks)
	require.NoError(t, err, "Failed to parse CSV with multiline descriptions")

	// We expect 2 tasks in this fixture
	require.Len(t, tasks, 2, "Should parse exactly 2 tasks")

	// First task has multiline content in both Title and Content fields
	task1 := tasks[0]
	assert.Equal(t, "Work", task1.FolderName)
	assert.Equal(t, "Project Alpha", task1.ProjectName)

	// The title contains a newline
	assert.Contains(t, task1.Title, "Task with multiline")
	assert.Contains(t, task1.Title, "description")
	assert.Contains(t, task1.Title, "\n", "Title should contain actual newline character")

	// The content contains multiple newlines and paragraphs
	assert.Contains(t, task1.Content, "This is a task description")
	assert.Contains(t, task1.Content, "that spans multiple lines")
	assert.Contains(t, task1.Content, "It has paragraphs and everything!")
	assert.Contains(t, task1.Content, "Including special characters: #, *, @")

	// Count newlines in content - should have at least 3 (between the 4 lines)
	newlineCount := strings.Count(task1.Content, "\n")
	assert.GreaterOrEqual(t, newlineCount, 3, "Content should have multiple newlines")

	// Second task is a regular task without multiline content
	task2 := tasks[1]
	assert.Equal(t, "Regular task", task2.Title)
	assert.Equal(t, "Simple description", task2.Content)
	assert.NotContains(t, task2.Title, "\n", "Regular task title should not have newlines")

	t.Logf("Successfully parsed tasks with multiline content:")
	t.Logf("  Task 1 title: %q", task1.Title)
	t.Logf("  Task 1 content: %q", task1.Content)
	t.Logf("  Task 2 title: %q", task2.Title)
}

func TestEmptyLabelHandlingWithRealCSV(t *testing.T) {
	t.Run("Parse CSV file", func(t *testing.T) {
		file, err := os.Open("testdata_ticktick_export.csv")
		require.NoError(t, err)
		defer file.Close()

		stat, err := file.Stat()
		require.NoError(t, err)

		lines, err := linesToSkipBeforeHeader(file, stat.Size())
		require.NoError(t, err)

		// Reset file position
		_, err = file.Seek(0, io.SeekStart)
		require.NoError(t, err)

		dec, err := newLineSkipDecoder(file, lines)
		require.NoError(t, err)
		tasks := []*tickTickTask{}
		err = gocsv.UnmarshalDecoder(dec, &tasks)
		require.NoError(t, err)
		require.Greater(t, len(tasks), 0)

		t.Logf("Successfully parsed %d tasks from CSV file", len(tasks))
	})

	t.Run("Process tags and check for empty labels", func(t *testing.T) {
		file, err := os.Open("testdata_ticktick_export.csv")
		require.NoError(t, err)
		defer file.Close()

		stat, err := file.Stat()
		require.NoError(t, err)

		lines, err := linesToSkipBeforeHeader(file, stat.Size())
		require.NoError(t, err)

		// Reset file position
		_, err = file.Seek(0, io.SeekStart)
		require.NoError(t, err)

		dec, err := newLineSkipDecoder(file, lines)
		require.NoError(t, err)
		tasks := []*tickTickTask{}
		err = gocsv.UnmarshalDecoder(dec, &tasks)
		require.NoError(t, err)

		// Process tags as the migration code does
		for _, task := range tasks {
			task.Tags = strings.Split(task.TagsList, ", ")
		}

		// Convert to Vikunja format
		vikunjaTasks := convertTickTickToVikunja(tasks)

		// Check all tasks for empty labels
		totalLabels := 0
		for _, project := range vikunjaTasks {
			for _, task := range project.Tasks {
				totalLabels += len(task.Labels)
				for _, label := range task.Labels {
					assert.NotEmpty(t, strings.TrimSpace(label.Title),
						"No label should be empty or whitespace-only. Found empty label in task: %s", task.Title)
				}
			}
		}

		t.Logf("Successfully processed %d tasks with %d total labels, no empty labels created", len(tasks), totalLabels)
	})
}
