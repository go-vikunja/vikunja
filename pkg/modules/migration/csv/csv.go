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
	"encoding/csv"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

// Migrator is the CSV migrator
type Migrator struct{}

// Name returns the name of this migrator
func (m *Migrator) Name() string {
	return "csv"
}

// SupportedDelimiters contains all supported CSV delimiters
var SupportedDelimiters = []string{",", ";", "\t", "|"}

// SupportedQuoteChars contains all supported quote characters
var SupportedQuoteChars = []string{"\"", "'"}

// SupportedDateFormats contains common date formats for parsing
var SupportedDateFormats = []string{
	"2006-01-02",                    // ISO date
	"2006-01-02T15:04:05",           // ISO datetime
	"2006-01-02T15:04:05Z07:00",     // RFC3339
	"2006-01-02T15:04:05-0700",      // ISO with timezone
	"02/01/2006",                    // DD/MM/YYYY
	"01/02/2006",                    // MM/DD/YYYY
	"02-01-2006",                    // DD-MM-YYYY
	"01-02-2006",                    // MM-DD-YYYY
	"Jan 2, 2006",                   // Month D, YYYY
	"2 Jan 2006",                    // D Month YYYY
	"02/01/2006 15:04",              // DD/MM/YYYY HH:MM
	"01/02/2006 15:04",              // MM/DD/YYYY HH:MM
	"2006-01-02 15:04:05",           // MySQL datetime
	"2006/01/02",                    // YYYY/MM/DD
	"02.01.2006",                    // DD.MM.YYYY (European)
	"02.01.2006 15:04",              // DD.MM.YYYY HH:MM (European)
	time.RFC1123,                    // RFC1123
	time.RFC1123Z,                   // RFC1123 with numeric zone
	time.RFC822,                     // RFC822
	time.RFC822Z,                    // RFC822 with numeric zone
	time.RFC850,                     // RFC850
	time.ANSIC,                      // ANSIC
	time.UnixDate,                   // Unix date
}

// TaskAttribute represents a task attribute that can be mapped from CSV
type TaskAttribute string

const (
	AttrTitle       TaskAttribute = "title"
	AttrDescription TaskAttribute = "description"
	AttrDueDate     TaskAttribute = "due_date"
	AttrStartDate   TaskAttribute = "start_date"
	AttrEndDate     TaskAttribute = "end_date"
	AttrDone        TaskAttribute = "done"
	AttrPriority    TaskAttribute = "priority"
	AttrLabels      TaskAttribute = "labels"
	AttrProject     TaskAttribute = "project"
	AttrReminder    TaskAttribute = "reminder"
	AttrIgnore      TaskAttribute = "ignore"
)

// AllTaskAttributes returns all available task attributes for mapping
var AllTaskAttributes = []TaskAttribute{
	AttrTitle,
	AttrDescription,
	AttrDueDate,
	AttrStartDate,
	AttrEndDate,
	AttrDone,
	AttrPriority,
	AttrLabels,
	AttrProject,
	AttrReminder,
	AttrIgnore,
}

// ColumnMapping represents a mapping from a CSV column to a task attribute
type ColumnMapping struct {
	ColumnIndex int           `json:"column_index"`
	ColumnName  string        `json:"column_name"`
	Attribute   TaskAttribute `json:"attribute"`
}

// DetectionResult contains the auto-detected CSV structure
type DetectionResult struct {
	Columns         []string `json:"columns"`
	Delimiter       string   `json:"delimiter"`
	QuoteChar       string   `json:"quote_char"`
	DateFormat      string   `json:"date_format"`
	SuggestedMapping []ColumnMapping `json:"suggested_mapping"`
	PreviewRows     [][]string `json:"preview_rows"`
}

// ImportConfig contains the configuration for CSV import
type ImportConfig struct {
	Delimiter   string          `json:"delimiter"`
	QuoteChar   string          `json:"quote_char"`
	DateFormat  string          `json:"date_format"`
	Mapping     []ColumnMapping `json:"mapping"`
}

// PreviewTask represents a task preview before import
type PreviewTask struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DueDate     string   `json:"due_date,omitempty"`
	StartDate   string   `json:"start_date,omitempty"`
	EndDate     string   `json:"end_date,omitempty"`
	Done        bool     `json:"done"`
	Priority    int      `json:"priority"`
	Labels      []string `json:"labels,omitempty"`
	Project     string   `json:"project,omitempty"`
}

// PreviewResult contains preview data before import
type PreviewResult struct {
	Tasks      []PreviewTask `json:"tasks"`
	TotalRows  int           `json:"total_rows"`
	ErrorCount int           `json:"error_count"`
	Errors     []string      `json:"errors,omitempty"`
}

// stripBOM removes the UTF-8 BOM from the beginning of a reader
func stripBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// detectDelimiter attempts to auto-detect the CSV delimiter
func detectDelimiter(data []byte) string {
	content := string(data)

	// Count occurrences of each delimiter in the first few lines
	lines := strings.SplitN(content, "\n", 5)
	if len(lines) < 2 {
		return "," // Default to comma
	}

	delimiterCounts := make(map[string]int)
	for _, delim := range SupportedDelimiters {
		count := 0
		for _, line := range lines[:min(3, len(lines))] {
			count += strings.Count(line, delim)
		}
		delimiterCounts[delim] = count
	}

	// Find the delimiter with the most consistent count across lines
	bestDelimiter := ","
	maxCount := 0
	for delim, count := range delimiterCounts {
		if count > maxCount {
			maxCount = count
			bestDelimiter = delim
		}
	}

	return bestDelimiter
}

// detectQuoteChar attempts to auto-detect the quote character
func detectQuoteChar(data []byte) string {
	content := string(data)

	doubleQuotes := strings.Count(content, "\"")
	singleQuotes := strings.Count(content, "'")

	if singleQuotes > doubleQuotes {
		return "'"
	}
	return "\""
}

// detectDateFormat attempts to detect the date format from sample data
func detectDateFormat(sampleDates []string) string {
	if len(sampleDates) == 0 {
		return SupportedDateFormats[0] // Default to ISO
	}

	for _, format := range SupportedDateFormats {
		matches := 0
		for _, dateStr := range sampleDates {
			dateStr = strings.TrimSpace(dateStr)
			if dateStr == "" {
				continue
			}
			_, err := time.Parse(format, dateStr)
			if err == nil {
				matches++
			}
		}
		// If most dates match this format, use it
		if matches > 0 && matches >= len(sampleDates)/2 {
			return format
		}
	}

	return SupportedDateFormats[0]
}

// suggestMapping suggests column mappings based on column names
func suggestMapping(columns []string) []ColumnMapping {
	mappings := make([]ColumnMapping, len(columns))

	// Common column name patterns for each attribute
	patterns := map[TaskAttribute][]string{
		AttrTitle:       {"title", "name", "task", "subject", "summary"},
		AttrDescription: {"description", "content", "notes", "details", "body", "text"},
		AttrDueDate:     {"due", "due_date", "duedate", "deadline", "due date"},
		AttrStartDate:   {"start", "start_date", "startdate", "begin", "start date"},
		AttrEndDate:     {"end", "end_date", "enddate", "finish", "end date"},
		AttrDone:        {"done", "completed", "complete", "finished", "status", "is_done"},
		AttrPriority:    {"priority", "importance", "urgent", "prio"},
		AttrLabels:      {"labels", "tags", "categories", "category", "label", "tag"},
		AttrProject:     {"project", "list", "folder", "group", "project_name", "list_name"},
		AttrReminder:    {"reminder", "remind", "alert", "notification"},
	}

	usedAttributes := make(map[TaskAttribute]bool)

	for i, col := range columns {
		colLower := strings.ToLower(strings.TrimSpace(col))
		mappings[i] = ColumnMapping{
			ColumnIndex: i,
			ColumnName:  col,
			Attribute:   AttrIgnore,
		}

		for attr, keywords := range patterns {
			if usedAttributes[attr] && attr != AttrLabels {
				continue // Don't map the same attribute twice (except labels)
			}
			for _, keyword := range keywords {
				if strings.Contains(colLower, keyword) || colLower == keyword {
					mappings[i].Attribute = attr
					usedAttributes[attr] = true
					break
				}
			}
			if mappings[i].Attribute != AttrIgnore {
				break
			}
		}
	}

	return mappings
}

// parseCSV parses CSV data with the given configuration
func parseCSV(data []byte, delimiter, quoteChar string) ([]string, [][]string, error) {
	data = stripBOM(data)
	reader := csv.NewReader(bytes.NewReader(data))

	if len(delimiter) > 0 {
		reader.Comma = rune(delimiter[0])
	}
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) == 0 {
		return nil, nil, &migration.ErrFileIsEmpty{}
	}

	headers := records[0]
	var dataRows [][]string
	if len(records) > 1 {
		dataRows = records[1:]
	}

	return headers, dataRows, nil
}

// DetectCSVStructure analyzes a CSV file and returns detection results
func DetectCSVStructure(file io.ReaderAt, size int64) (*DetectionResult, error) {
	if size == 0 {
		return nil, &migration.ErrFileIsEmpty{}
	}

	// Read the entire file
	data := make([]byte, size)
	_, err := file.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Detect delimiter and quote character
	delimiter := detectDelimiter(data)
	quoteChar := detectQuoteChar(data)

	// Parse CSV
	headers, rows, err := parseCSV(data, delimiter, quoteChar)
	if err != nil {
		return nil, &migration.ErrNotACSVFile{}
	}

	// Suggest column mappings
	suggestedMapping := suggestMapping(headers)

	// Collect sample dates for format detection
	var sampleDates []string
	for _, mapping := range suggestedMapping {
		if mapping.Attribute == AttrDueDate || mapping.Attribute == AttrStartDate || mapping.Attribute == AttrEndDate {
			for _, row := range rows {
				if mapping.ColumnIndex < len(row) && row[mapping.ColumnIndex] != "" {
					sampleDates = append(sampleDates, row[mapping.ColumnIndex])
					if len(sampleDates) >= 10 {
						break
					}
				}
			}
		}
	}

	dateFormat := detectDateFormat(sampleDates)

	// Get preview rows (first 5)
	previewRows := rows
	if len(previewRows) > 5 {
		previewRows = previewRows[:5]
	}

	return &DetectionResult{
		Columns:          headers,
		Delimiter:        delimiter,
		QuoteChar:        quoteChar,
		DateFormat:       dateFormat,
		SuggestedMapping: suggestedMapping,
		PreviewRows:      previewRows,
	}, nil
}

// PreviewImport generates a preview of the import based on current mapping
func PreviewImport(file io.ReaderAt, size int64, config ImportConfig) (*PreviewResult, error) {
	if size == 0 {
		return nil, &migration.ErrFileIsEmpty{}
	}

	data := make([]byte, size)
	_, err := file.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	_, rows, err := parseCSV(data, config.Delimiter, config.QuoteChar)
	if err != nil {
		return nil, &migration.ErrNotACSVFile{}
	}

	result := &PreviewResult{
		Tasks:     make([]PreviewTask, 0, min(5, len(rows))),
		TotalRows: len(rows),
	}

	previewCount := min(5, len(rows))
	for i := 0; i < previewCount; i++ {
		task, err := rowToPreviewTask(rows[i], config)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.Tasks = append(result.Tasks, task)
	}

	return result, nil
}

// rowToPreviewTask converts a CSV row to a preview task
func rowToPreviewTask(row []string, config ImportConfig) (PreviewTask, error) {
	task := PreviewTask{}

	for _, mapping := range config.Mapping {
		if mapping.ColumnIndex >= len(row) {
			continue
		}

		value := strings.TrimSpace(row[mapping.ColumnIndex])
		if value == "" {
			continue
		}

		switch mapping.Attribute {
		case AttrTitle:
			task.Title = value
		case AttrDescription:
			task.Description = value
		case AttrDueDate:
			task.DueDate = value
		case AttrStartDate:
			task.StartDate = value
		case AttrEndDate:
			task.EndDate = value
		case AttrDone:
			task.Done = parseBool(value)
		case AttrPriority:
			task.Priority = parsePriority(value)
		case AttrLabels:
			task.Labels = parseLabels(value)
		case AttrProject:
			task.Project = value
		}
	}

	return task, nil
}

// parseBool parses various boolean representations
func parseBool(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "true" || lower == "yes" || lower == "1" || lower == "done" || lower == "completed"
}

// parsePriority parses priority value
func parsePriority(value string) int {
	// Try to parse as number
	if p, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		// Vikunja uses 0-5 priority (0=unset, 1=low, 5=urgent)
		if p < 0 {
			return 0
		}
		if p > 5 {
			return 5
		}
		return p
	}

	// Try to parse text priority
	lower := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(lower, "urgent") || strings.Contains(lower, "highest"):
		return 5
	case strings.Contains(lower, "high"):
		return 4
	case strings.Contains(lower, "medium") || strings.Contains(lower, "normal"):
		return 3
	case strings.Contains(lower, "low"):
		return 2
	case strings.Contains(lower, "lowest"):
		return 1
	}

	return 0
}

// parseLabels parses comma-separated labels
func parseLabels(value string) []string {
	parts := strings.Split(value, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		label := strings.TrimSpace(part)
		if label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

// parseDate parses a date string with the given format
func parseDate(value, format string) time.Time {
	if value == "" {
		return time.Time{}
	}

	// Try the specified format first
	if t, err := time.Parse(format, strings.TrimSpace(value)); err == nil {
		return t
	}

	// Try all supported formats as fallback
	for _, f := range SupportedDateFormats {
		if t, err := time.Parse(f, strings.TrimSpace(value)); err == nil {
			return t
		}
	}

	return time.Time{}
}

// Migrate imports CSV data into Vikunja
// @Summary Import all tasks from a CSV file
// @Description Imports tasks from a CSV file into Vikunja. Requires a mapping configuration.
// @tags migration
// @Accept multipart/form-data
// @Produce json
// @Security JWTKeyAuth
// @Param import formData file true "The CSV file to import"
// @Param config formData string true "The import configuration JSON"
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 400 {object} models.Message "Invalid CSV file or configuration"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/csv/migrate [put]
func (m *Migrator) Migrate(u *user.User, file io.ReaderAt, size int64) error {
	// This will be called with the standard file migrator handler
	// The actual configuration will come through the handler
	return &migration.ErrNotACSVFile{} // Need config, use MigrateWithConfig instead
}

// MigrateWithConfig imports CSV data into Vikunja with the provided configuration
func MigrateWithConfig(u *user.User, file io.ReaderAt, size int64, config ImportConfig) error {
	if size == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	data := make([]byte, size)
	_, err := file.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return err
	}

	_, rows, err := parseCSV(data, config.Delimiter, config.QuoteChar)
	if err != nil {
		return &migration.ErrNotACSVFile{}
	}

	if len(rows) == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	// Convert rows to Vikunja structure
	vikunjaTasks := convertToVikunja(rows, config)

	return migration.InsertFromStructure(vikunjaTasks, u)
}

// convertToVikunja converts CSV rows to Vikunja project/task structure
func convertToVikunja(rows [][]string, config ImportConfig) []*models.ProjectWithTasksAndBuckets {
	var pseudoParentID int64 = 1
	result := []*models.ProjectWithTasksAndBuckets{
		{
			Project: models.Project{
				ID:    pseudoParentID,
				Title: "Imported from CSV",
			},
		},
	}

	projects := make(map[string]*models.ProjectWithTasksAndBuckets)
	defaultProjectName := "Tasks"

	for i, row := range rows {
		task := rowToTask(row, config, int64(i+1))

		// Determine project name
		projectName := defaultProjectName
		for _, mapping := range config.Mapping {
			if mapping.Attribute == AttrProject && mapping.ColumnIndex < len(row) {
				if pn := strings.TrimSpace(row[mapping.ColumnIndex]); pn != "" {
					projectName = pn
				}
			}
		}

		// Get or create project
		if _, exists := projects[projectName]; !exists {
			projects[projectName] = &models.ProjectWithTasksAndBuckets{
				Project: models.Project{
					ID:              int64(len(projects)+2) + pseudoParentID,
					ParentProjectID: pseudoParentID,
					Title:           projectName,
				},
			}
		}

		// Add task to project
		projects[projectName].Tasks = append(projects[projectName].Tasks, &models.TaskWithComments{Task: task})
	}

	// Collect all projects
	for _, p := range projects {
		result = append(result, p)
	}

	// Sort projects by title for consistent ordering
	sort.Slice(result[1:], func(i, j int) bool {
		return result[i+1].Title < result[j+1].Title
	})

	return result
}

// rowToTask converts a CSV row to a Vikunja task
func rowToTask(row []string, config ImportConfig, taskID int64) models.Task {
	task := models.Task{
		ID: taskID,
	}

	for _, mapping := range config.Mapping {
		if mapping.ColumnIndex >= len(row) {
			continue
		}

		value := strings.TrimSpace(row[mapping.ColumnIndex])
		if value == "" {
			continue
		}

		switch mapping.Attribute {
		case AttrTitle:
			task.Title = value
		case AttrDescription:
			task.Description = value
		case AttrDueDate:
			task.DueDate = parseDate(value, config.DateFormat)
		case AttrStartDate:
			task.StartDate = parseDate(value, config.DateFormat)
		case AttrEndDate:
			task.EndDate = parseDate(value, config.DateFormat)
		case AttrDone:
			task.Done = parseBool(value)
			if task.Done {
				task.DoneAt = time.Now()
			}
		case AttrPriority:
			task.Priority = int64(parsePriority(value))
		case AttrLabels:
			labels := parseLabels(value)
			for _, labelTitle := range labels {
				task.Labels = append(task.Labels, &models.Label{Title: labelTitle})
			}
		case AttrReminder:
			// Parse reminder as duration or date
			reminderDate := parseDate(value, config.DateFormat)
			if !reminderDate.IsZero() {
				task.Reminders = []*models.TaskReminder{
					{
						Reminder: reminderDate,
					},
				}
			}
		}
	}

	// Ensure task has a title
	if task.Title == "" {
		task.Title = "Untitled Task"
	}

	return task
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
