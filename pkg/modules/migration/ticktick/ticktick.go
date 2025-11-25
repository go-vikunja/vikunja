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
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"github.com/gocarina/gocsv"
)

const timeISO = "2006-01-02T15:04:05-0700"

type Migrator struct {
}

type tickTickTask struct {
	FolderName        string        `csv:"Folder Name"`
	ProjectName       string        `csv:"List Name"`
	Title             string        `csv:"Title"`
	TagsList          string        `csv:"Tags"`
	Tags              []string      `csv:"-"`
	Content           string        `csv:"Content"`
	IsChecklistString string        `csv:"Is Check list"`
	IsChecklist       bool          `csv:"-"`
	StartDate         tickTickTime  `csv:"Start Date"`
	DueDate           tickTickTime  `csv:"Due Date"`
	ReminderDuration  string        `csv:"Reminder"`
	Reminder          time.Duration `csv:"-"`
	Repeat            string        `csv:"Repeat"`
	Priority          int           `csv:"Priority"`
	Status            string        `csv:"Status"`
	CreatedTime       tickTickTime  `csv:"Created Time"`
	CompletedTime     tickTickTime  `csv:"Completed Time"`
	Order             float64       `csv:"Order"`
	TaskID            int64         `csv:"taskId"`
	ParentID          int64         `csv:"parentId"`
}

type tickTickTime struct {
	time.Time
}

func (date *tickTickTime) UnmarshalCSV(csv string) (err error) {
	date.Time = time.Time{}
	if csv == "" {
		return nil
	}
	date.Time, err = time.Parse(timeISO, csv)
	return err
}

func convertTickTickToVikunja(tasks []*tickTickTask) (result []*models.ProjectWithTasksAndBuckets) {
	var pseudoParentID int64 = 1
	result = []*models.ProjectWithTasksAndBuckets{
		{
			Project: models.Project{
				ID:    pseudoParentID,
				Title: "Migrated from TickTick",
			},
		},
	}

	projects := make(map[string]*models.ProjectWithTasksAndBuckets)
	for index, t := range tasks {
		_, has := projects[t.ProjectName]
		if !has {
			projects[t.ProjectName] = &models.ProjectWithTasksAndBuckets{
				Project: models.Project{
					ID:              int64(index+1) + pseudoParentID,
					ParentProjectID: pseudoParentID,
					Title:           t.ProjectName,
				},
			}
		}

		labels := make([]*models.Label, 0, len(t.Tags))
		for _, tag := range t.Tags {
			// Only create labels for non-empty tags after trimming whitespace
			trimmedTag := strings.TrimSpace(tag)
			if trimmedTag != "" {
				labels = append(labels, &models.Label{
					Title: trimmedTag,
				})
			}
		}

		task := &models.TaskWithComments{
			Task: models.Task{
				ID:          t.TaskID,
				Title:       t.Title,
				Description: t.Content,
				StartDate:   t.StartDate.Time,
				EndDate:     t.DueDate.Time,
				DueDate:     t.DueDate.Time,
				Done:        t.Status == "1",
				DoneAt:      t.CompletedTime.Time,
				Position:    t.Order,
				Labels:      labels,
			},
		}

		if !t.DueDate.IsZero() && t.Reminder > 0 {
			task.Reminders = []*models.TaskReminder{
				{
					RelativeTo:     models.ReminderRelationDueDate,
					RelativePeriod: int64((t.Reminder * -1).Seconds()),
				},
			}
		}

		if t.ParentID != 0 {
			task.RelatedTasks = map[models.RelationKind][]*models.Task{
				models.RelationKindParenttask: {{ID: t.ParentID}},
			}
		}

		projects[t.ProjectName].Tasks = append(projects[t.ProjectName].Tasks, task)
	}

	for _, l := range projects {
		result = append(result, l)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})

	return
}

// Name is used to get the name of the ticktick migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/status [get]
func (m *Migrator) Name() string {
	return "ticktick"
}

// stripBOM removes the UTF-8 BOM from the beginning of a reader
func stripBOM(r io.Reader) io.Reader {
	// Read the first few bytes to check for BOM
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		// If we read some bytes before the error, preserve them
		if n > 0 {
			return io.MultiReader(bytes.NewReader(buf[:n]), r)
		}
		return r
	}

	// Check if it starts with UTF-8 BOM (0xEF, 0xBB, 0xBF)
	// We need exactly 3 bytes and they must match the BOM sequence
	if n == 3 && len(buf) >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		// BOM found, return reader without BOM
		return io.MultiReader(bytes.NewReader(buf[3:n]), r)
	}

	// No BOM found, return reader with the bytes we read back
	return io.MultiReader(bytes.NewReader(buf[:n]), r)
}

func newLineSkipDecoder(r io.Reader, linesToSkip int) (gocsv.SimpleDecoder, error) {
	// Strip BOM if present - this must be done consistently with linesToSkipBeforeHeader
	r = stripBOM(r)

	// Read all content into memory so we can work with it
	// This is acceptable since CSV imports are typically not huge files
	allBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Skip the metadata lines before the CSV header by finding byte offsets
	// We use a scanner here because these metadata lines are plain text, not CSV data,
	// so they won't contain multiline quoted fields
	scanner := bufio.NewScanner(bytes.NewReader(allBytes))
	bytesSkipped := 0
	for i := 0; i < linesToSkip; i++ {
		if !scanner.Scan() {
			break
		}
		// Count bytes for this line plus the newline character
		bytesSkipped += len(scanner.Bytes()) + 1 // +1 for the \n
	}

	// Check for errors after skipping lines
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Now create a CSV reader starting from after the skipped lines
	// This preserves any multiline quoted fields in the CSV data
	remainingContent := allBytes[bytesSkipped:]
	reader := csv.NewReader(bytes.NewReader(remainingContent))

	// Allow variable field counts and be lenient with parsing
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return gocsv.NewSimpleDecoderFromCSVReader(reader), nil
}

func linesToSkipBeforeHeader(file io.ReaderAt, size int64) (int, error) {
	sr := io.NewSectionReader(file, 0, size)
	// Strip BOM before scanning for header
	r := stripBOM(sr)
	scanner := bufio.NewScanner(r)
	lines := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Folder Name") &&
			strings.Contains(line, "List Name") &&
			strings.Contains(line, "Title") {
			break
		}
		lines++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return lines, nil
}

// Migrate takes a ticktick export, parses it and imports everything in it into Vikunja.
// @Summary Import all projects, tasks etc. from a TickTick backup export
// @Description Imports all projects, tasks, notes, reminders, subtasks and files from a TickTick backup export into Vikunja.
// @tags migration
// @Accept x-www-form-urlencoded
// @Produce json
// @Security JWTKeyAuth
// @Param import formData string true "The TickTick backup csv file."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/migrate [post]
func (m *Migrator) Migrate(user *user.User, file io.ReaderAt, size int64) error {
	// Check if file is empty
	if size == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	fr := io.NewSectionReader(file, 0, size)

	// Check if the file is a valid CSV
	buf := make([]byte, 1024)
	n, err := fr.Read(buf)
	if errors.Is(err, io.EOF) || n == 0 {
		return &migration.ErrFileIsEmpty{}
	}
	if err != nil {
		return err
	}

	// Reset the reader position to start
	_, err = fr.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Check if the content looks like a CSV file
	content := string(buf[:n])
	if !isValidCSV(content) {
		return &migration.ErrNotACSVFile{}
	}

	allTasks := []*tickTickTask{}
	skip, err := linesToSkipBeforeHeader(file, size)
	if err != nil {
		return err
	}
	decode, err := newLineSkipDecoder(fr, skip)
	if err != nil {
		return err
	}
	err = gocsv.UnmarshalDecoder(decode, &allTasks)
	if err != nil {
		return err
	}

	// Also check if no tasks were found after decoding
	if len(allTasks) == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	for _, task := range allTasks {
		if task.IsChecklistString == "Y" {
			task.IsChecklist = true
		}

		reminder := utils.ParseISO8601Duration(task.ReminderDuration)
		if reminder > 0 {
			task.Reminder = reminder
		}

		task.Tags = strings.Split(task.TagsList, ", ")
	}

	vikunjaTasks := convertTickTickToVikunja(allTasks)

	return migration.InsertFromStructure(vikunjaTasks, user)
}

// isValidCSV performs a basic check to determine if the content looks like a CSV file
func isValidCSV(content string) bool {
	// Check for common CSV headers from TickTick export
	if !strings.Contains(content, "Folder Name") ||
		!strings.Contains(content, "List Name") ||
		!strings.Contains(content, "Title") {
		return false
	}

	// Check if the file has commas as separators and multiple lines
	hasCommas := strings.Contains(content, ",")
	hasNewlines := strings.Contains(content, "\n")

	return hasCommas && hasNewlines
}
