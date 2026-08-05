// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package vikunjafile

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration/focalboard"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func writeJSONL(t *testing.T, writer *zip.Writer, name string, records []map[string]any) {
	t.Helper()
	entry, err := writer.Create(name)
	require.NoError(t, err)
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	for _, record := range records {
		require.NoError(t, encoder.Encode(record))
	}
	_, err = entry.Write(data.Bytes())
	require.NoError(t, err)
}

func sourceBoard(id, title string, options []any) map[string]any {
	return map[string]any{"type": "board", "data": map[string]any{
		"id": id, "type": "P", "title": title, "createAt": 1000, "updateAt": 2000, "deleteAt": 0, "isTemplate": false,
		"cardProperties": []any{
			map[string]any{"id": "status", "name": "Status", "type": "select", "options": options},
			map[string]any{"id": "assignee", "name": "Кто делает", "type": "text", "options": []any{}},
			map[string]any{"id": "priority", "name": "Приоритет", "type": "text", "options": []any{}},
			map[string]any{"id": "due", "name": "Срок", "type": "text", "options": []any{}},
		},
	}}
}

func sourceView(id, boardID string) map[string]any {
	return map[string]any{"type": "block", "data": map[string]any{"id": id, "type": "view", "boardId": boardID, "parentId": boardID, "title": "Kanban", "createAt": 1000, "updateAt": 2000, "deleteAt": 0, "fields": map[string]any{"viewType": "board"}}}
}
func sourceCard(id, boardID, title, status string, created int64, content []any, priority, due string) map[string]any {
	return map[string]any{"type": "block", "data": map[string]any{"id": id, "type": "card", "boardId": boardID, "parentId": boardID, "title": title, "createAt": created, "updateAt": created + 10, "deleteAt": 0, "fields": map[string]any{"contentOrder": content, "isTemplate": false, "properties": map[string]any{"status": status, "priority": priority, "due": due, "assignee": ""}}}}
}
func sourceText(id, boardID, text string, created int64) map[string]any {
	return map[string]any{"type": "block", "data": map[string]any{"id": id, "type": "text", "boardId": boardID, "parentId": boardID, "title": text, "createAt": created, "updateAt": created + 10, "deleteAt": 0, "fields": map[string]any{}}}
}

func syntheticFocalboardPackage(t *testing.T) string {
	t.Helper()
	standard := []any{map[string]any{"id": "todo", "value": "Todo"}, map[string]any{"id": "doing", "value": "In Progress"}, map[string]any{"id": "done", "value": "Done"}}
	backlog := []any{map[string]any{"id": "new", "value": "Новая"}, map[string]any{"id": "working", "value": "В работе"}, map[string]any{"id": "sorted", "value": "Разобрана"}}
	boardOne := []map[string]any{
		sourceBoard("board-1", "Synthetic Standard", standard), sourceView("view-1", "board-1"),
		sourceCard("card-direct", "board-1", "Duplicate title", "todo", 3000, []any{"text-direct"}, "🔥 P0 Срочно + важно", "2026-08-04"),
		sourceText("text-direct", "board-1", "Direct description", 3100),
		sourceCard("card-recovered", "board-1", "Duplicate title", "done", 5000, []any{"missing-text"}, "🧊 P3 Backlog", ""),
		sourceText("text-recovered", "board-1", "Recovered description", 5600),
	}
	boardTwo := []map[string]any{
		sourceBoard("board-2", "Synthetic Backlog", backlog), sourceView("view-2", "board-2"),
		sourceCard("card-new", "board-2", "Backlog task", "new", 7000, []any{}, "", ""),
	}
	var nested bytes.Buffer
	nw := zip.NewWriter(&nested)
	entry, err := nw.Create("version.json")
	require.NoError(t, err)
	_, err = entry.Write([]byte(`{"version":2,"date":1722686400000}`))
	require.NoError(t, err)
	writeJSONL(t, nw, "board-1/board.jsonl", boardOne)
	writeJSONL(t, nw, "board-2/board.jsonl", boardTwo)
	require.NoError(t, nw.Close())
	var outer bytes.Buffer
	ow := zip.NewWriter(&outer)
	entry, err = ow.Create("package/archive.boardarchive")
	require.NoError(t, err)
	_, err = entry.Write(nested.Bytes())
	require.NoError(t, err)
	require.NoError(t, ow.Close())
	path := filepath.Join(t.TempDir(), "focalboard-package.zip")
	require.NoError(t, os.WriteFile(path, outer.Bytes(), 0o600))
	return path
}

func TestGeneratedFocalboardZipImportsWithCurrentMigrator(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	input := syntheticFocalboardPackage(t)
	result, err := focalboard.Convert(input, focalboard.Options{Strict: true, Timezone: "Europe/Moscow", VikunjaVersion: focalboard.DefaultVikunjaVersion})
	require.NoError(t, err)
	require.Len(t, result.Normalized.Boards, 2)
	require.Equal(t, 3, result.Normalized.Counts.Cards)
	require.Equal(t, 1, result.Normalized.Counts.DirectDescriptions)
	require.Equal(t, 1, result.Normalized.Counts.RecoveredDescriptions)
	require.Equal(t, 1, result.Normalized.Counts.DuplicateTitleExtras)
	require.Equal(t, []int64{1, 2, 3}, []int64{result.Reconciliation.CardMappings[0].TargetTaskLegacyID, result.Reconciliation.CardMappings[1].TargetTaskLegacyID, result.Reconciliation.CardMappings[2].TargetTaskLegacyID})
	require.Equal(t, "Done", result.Normalized.Boards[0].Tasks[1].StatusRaw)
	zr, zipErr := zip.NewReader(bytes.NewReader(result.VikunjaZip), int64(len(result.VikunjaZip)))
	require.NoError(t, zipErr)
	var exported []*models.ProjectWithTasksAndBuckets
	for _, entry := range zr.File {
		if entry.Name == "data.json" {
			r, openErr := entry.Open()
			require.NoError(t, openErr)
			require.NoError(t, json.NewDecoder(r).Decode(&exported))
			require.NoError(t, r.Close())
		}
	}
	require.False(t, exported[1].Tasks[1].Done)
	reader := bytes.NewReader(result.VikunjaZip)
	err = (&FileMigrator{}).Migrate(&user.User{ID: 1}, reader, int64(reader.Len()))
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	root := &models.Project{}
	exists, err := s.Where("title = ? AND owner_id = ?", focalboard.RootProjectTitle, int64(1)).Get(root)
	require.NoError(t, err)
	require.True(t, exists)
	standard := &models.Project{}
	exists, err = s.Where("title = ? AND owner_id = ?", "Synthetic Standard", int64(1)).Get(standard)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, root.ID, *standard.ParentProjectID)
	backlog := &models.Project{}
	exists, err = s.Where("title = ? AND owner_id = ?", "Synthetic Backlog", int64(1)).Get(backlog)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, root.ID, *backlog.ParentProjectID)

	var duplicateTasks []*models.Task
	require.NoError(t, s.Where("project_id = ? AND title = ?", standard.ID, "Duplicate title").Find(&duplicateTasks))
	require.Len(t, duplicateTasks, 2)
	moscow, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)
	methods := map[string]bool{}
	doneCount := 0
	dueCount := 0
	for _, task := range duplicateTasks {
		require.Contains(t, task.Description, "<!-- focalboard-migration:")
		require.Contains(t, task.Description, "source_card_id")
		if task.Done {
			doneCount++
		}
		if !task.DueDate.IsZero() {
			dueCount++
			local := task.DueDate.In(moscow)
			require.Equal(t, 2026, local.Year())
			require.Equal(t, time.August, local.Month())
			require.Equal(t, 4, local.Day())
			require.Equal(t, int64(5), task.Priority)
		}
		if strings.Contains(task.Description, "Direct description") {
			methods["direct"] = true
		}
		if strings.Contains(task.Description, "Recovered description") {
			methods["recovered"] = true
		}
	}
	require.Equal(t, map[string]bool{"direct": true, "recovered": true}, methods)
	standardView := &models.ProjectView{}
	exists, err = s.Where("project_id = ? AND view_kind = ?", standard.ID, models.ProjectViewKindKanban).Get(standardView)
	require.NoError(t, err)
	require.True(t, exists)
	require.NotZero(t, standardView.DoneBucketID)
	doneBucket := &models.Bucket{}
	exists, err = s.ID(standardView.DoneBucketID).Get(doneBucket)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "Done", doneBucket.Title)
	var recoveredTask *models.Task
	for _, task := range duplicateTasks {
		if strings.Contains(task.Description, "Recovered description") {
			recoveredTask = task
		}
	}
	require.NotNil(t, recoveredTask)
	taskBucket := &models.TaskBucket{}
	exists, err = s.Where("task_id = ? AND project_view_id = ?", recoveredTask.ID, standardView.ID).Get(taskBucket)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, standardView.DoneBucketID, taskBucket.BucketID)
	require.Equal(t, 1, doneCount)
	require.Equal(t, 1, dueCount)
	for _, tc := range []struct {
		projectID int64
		title     string
	}{{standard.ID, "Todo"}, {standard.ID, "Done"}, {backlog.ID, "Новая"}, {backlog.ID, "В работе"}, {backlog.ID, "Разобрана"}} {
		view := &models.ProjectView{}
		exists, err = s.Where("project_id = ? AND view_kind = ?", tc.projectID, models.ProjectViewKindKanban).Get(view)
		require.NoError(t, err)
		require.True(t, exists)
		bucket := &models.Bucket{}
		exists, err = s.Where("project_view_id = ? AND title = ?", view.ID, tc.title).Get(bucket)
		require.NoError(t, err)
		require.True(t, exists)
	}
	db.AssertExists(t, "tasks", map[string]any{"project_id": backlog.ID, "title": "Backlog task", "created_by_id": int64(1)}, false)
}
