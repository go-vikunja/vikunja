// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package focalboard

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
)

type fixtureRecord struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func writeSyntheticPackage(t *testing.T, records []fixtureRecord, extraNested map[string][]byte) string {
	return writeSyntheticPackageWithVersion(t, records, extraNested, []byte(`{"version":2,"date":1722686400000}`), true)
}

func writeSyntheticPackageWithVersion(t *testing.T, records []fixtureRecord, extraNested map[string][]byte, versionData []byte, includeVersion bool) string {
	t.Helper()
	var jsonl bytes.Buffer
	enc := json.NewEncoder(&jsonl)
	for _, record := range records {
		require.NoError(t, enc.Encode(record))
	}
	var nested bytes.Buffer
	nw := zip.NewWriter(&nested)
	if includeVersion {
		w, err := nw.Create("version.json")
		require.NoError(t, err)
		_, err = w.Write(versionData)
		require.NoError(t, err)
	}
	w, err := nw.Create("board-1/board.jsonl")
	require.NoError(t, err)
	_, err = w.Write(jsonl.Bytes())
	require.NoError(t, err)
	for name, data := range extraNested {
		w, err = nw.Create(name)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, nw.Close())
	var outer bytes.Buffer
	ow := zip.NewWriter(&outer)
	w, err = ow.Create("package/archive.boardarchive")
	require.NoError(t, err)
	_, err = w.Write(nested.Bytes())
	require.NoError(t, err)
	require.NoError(t, ow.Close())
	packagePath := filepath.Join(t.TempDir(), "package.zip")
	require.NoError(t, os.WriteFile(packagePath, outer.Bytes(), 0o600))
	return packagePath
}

func baseBoardRecords(cardsAndTexts ...fixtureRecord) []fixtureRecord {
	board := fixtureRecord{Type: "board", Data: map[string]any{
		"id": "board-1", "type": "P", "title": "Synthetic board", "createAt": 1000, "updateAt": 2000,
		"deleteAt": 0, "isTemplate": false,
		"cardProperties": []any{
			map[string]any{"id": "status", "name": "Status", "type": "select", "options": []any{
				map[string]any{"id": "todo", "value": "Todo", "color": "default"},
				map[string]any{"id": "doing", "value": "In Progress", "color": "default"},
				map[string]any{"id": "done", "value": "Done", "color": "default"},
			}},
			map[string]any{"id": "assignee", "name": "Кто делает", "type": "text", "options": []any{}},
			map[string]any{"id": "priority", "name": "Приоритет", "type": "text", "options": []any{}},
			map[string]any{"id": "due", "name": "Срок", "type": "text", "options": []any{}},
		},
	}}
	view := fixtureRecord{Type: "block", Data: map[string]any{
		"id": "view-1", "type": "view", "boardId": "board-1", "parentId": "board-1",
		"title": "Kanban", "createAt": 1000, "updateAt": 2000, "deleteAt": 0,
		"fields": map[string]any{"viewType": "board"},
	}}
	return append([]fixtureRecord{board, view}, cardsAndTexts...)
}

func cardRecord(id, title string, created int, contentOrder []any, priority, due, assignee string) fixtureRecord {
	return fixtureRecord{Type: "block", Data: map[string]any{
		"id": id, "type": "card", "boardId": "board-1", "parentId": "board-1", "title": title,
		"createAt": created, "updateAt": created + 10, "deleteAt": 0,
		"fields": map[string]any{"contentOrder": contentOrder, "isTemplate": false, "properties": map[string]any{
			"status": "todo", "priority": priority, "due": due, "assignee": assignee,
		}},
	}}
}

func textRecord(id string, created int, description string) fixtureRecord {
	return fixtureRecord{Type: "block", Data: map[string]any{
		"id": id, "type": "text", "boardId": "board-1", "parentId": "board-1", "title": description,
		"createAt": created, "updateAt": created + 10, "deleteAt": 0, "fields": map[string]any{},
	}}
}

func TestConvertDirectRecoveredAndRawPreservation(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card-direct", "Same title", 10_000, []any{"text-direct"}, "🔥 P0 Срочно + важно", "2026-08-04", "Alice"),
		textRecord("text-direct", 10_100, "<p>direct</p>"),
		cardRecord("card-recovered", "Same title", 20_000, []any{"missing-text"}, "🧊 P3 Backlog", "after meeting", "Alice"),
		textRecord("text-orphan", 20_600, "<p>recovered</p>"),
		cardRecord("card-none", "No description", 30_000, []any{}, "", "", ""),
	), nil)

	result, err := Convert(input, Options{Strict: true, Timezone: "Europe/Moscow", VikunjaVersion: "2.5.0"})
	require.NoError(t, err)
	require.Len(t, result.Normalized.Boards, 1)
	require.Len(t, result.Normalized.Boards[0].Tasks, 3)
	require.Equal(t, []string{"view-1"}, result.Normalized.Boards[0].SourceViewIDs)

	tasks := map[string]NormalizedTask{}
	for _, task := range result.Normalized.Boards[0].Tasks {
		tasks[task.SourceCardID] = task
	}
	require.Equal(t, DescriptionDirect, tasks["card-direct"].DescriptionLinkMethod)
	require.Equal(t, "text-direct", tasks["card-direct"].SourceTextID)
	require.Equal(t, DescriptionRecovered, tasks["card-recovered"].DescriptionLinkMethod)
	require.Equal(t, "text-orphan", tasks["card-recovered"].SourceTextID)
	require.Equal(t, DescriptionNone, tasks["card-none"].DescriptionLinkMethod)
	require.EqualValues(t, 5, tasks["card-direct"].NativePriority)
	require.EqualValues(t, 1, tasks["card-recovered"].NativePriority)
	require.Equal(t, "after meeting", tasks["card-recovered"].DueRaw)
	require.Empty(t, tasks["card-recovered"].NativeDueDate)
	due, err := time.Parse(time.RFC3339, tasks["card-direct"].NativeDueDate)
	require.NoError(t, err)
	require.Equal(t, 8, int(due.Month()))
	require.Equal(t, 4, due.Day())
	require.Equal(t, "+03:00", due.Format("Z07:00"))
	require.Equal(t, "Alice", tasks["card-direct"].AssigneeRaw)
	require.Len(t, result.Reconciliation.RecoveredDescriptions, 1)
	require.Equal(t, int64(600), result.Reconciliation.RecoveredDescriptions[0].DeltaMS)
	require.Equal(t, "text-direct", result.Reconciliation.CardMappings[0].SourceTextID)
	require.NotEmpty(t, result.Reconciliation.CardMappings[0].SourceCreatedAt)
	reconciliationRows, err := csv.NewReader(bytes.NewReader(result.ReconciliationCSV)).ReadAll()
	require.NoError(t, err)
	require.Len(t, reconciliationRows, 4)
	for _, row := range reconciliationRows {
		require.Len(t, row, 11)
	}

	zr, err := zip.NewReader(bytes.NewReader(result.VikunjaZip), int64(len(result.VikunjaZip)))
	require.NoError(t, err)
	entries := map[string]bool{}
	var vikunjaData []byte
	for _, f := range zr.File {
		entries[f.Name] = true
		if f.Name == "data.json" {
			r, openErr := f.Open()
			require.NoError(t, openErr)
			vikunjaData, err = io.ReadAll(r)
			require.NoError(t, err)
			require.NoError(t, r.Close())
		}
	}
	require.True(t, entries["VERSION"])
	require.True(t, entries["data.json"])
	require.True(t, entries["filters.json"])
	var projects []*models.ProjectWithTasksAndBuckets
	require.NoError(t, json.Unmarshal(vikunjaData, &projects))
	require.Len(t, projects, 2)
	require.Len(t, projects[1].Tasks, 3)
	importedDirect := projects[1].Tasks[0]
	require.Equal(t, "2026-08-04T00:00:00+03:00", importedDirect.DueDate.Format(time.RFC3339))
	require.Contains(t, importedDirect.Description, `"source_text_id":"text-direct"`)
}

func writeRawSyntheticPackage(t *testing.T, boardJSONL []byte) string {
	t.Helper()
	var nested bytes.Buffer
	nw := zip.NewWriter(&nested)
	w, err := nw.Create("version.json")
	require.NoError(t, err)
	_, err = w.Write([]byte(`{"version":2,"date":1722686400000}`))
	require.NoError(t, err)
	w, err = nw.Create("board-1/board.jsonl")
	require.NoError(t, err)
	_, err = w.Write(boardJSONL)
	require.NoError(t, err)
	require.NoError(t, nw.Close())

	var outer bytes.Buffer
	ow := zip.NewWriter(&outer)
	w, err = ow.Create("package/archive.boardarchive")
	require.NoError(t, err)
	_, err = w.Write(nested.Bytes())
	require.NoError(t, err)
	require.NoError(t, ow.Close())
	path := filepath.Join(t.TempDir(), "raw-package.zip")
	require.NoError(t, os.WriteFile(path, outer.Bytes(), 0o600))
	return path
}

func writeMalformedNestedPackage(t *testing.T) string {
	t.Helper()
	var outer bytes.Buffer
	ow := zip.NewWriter(&outer)
	w, err := ow.Create("package/archive.boardarchive")
	require.NoError(t, err)
	_, err = w.Write([]byte("not-a-zip"))
	require.NoError(t, err)
	require.NoError(t, ow.Close())
	path := filepath.Join(t.TempDir(), "malformed-nested-package.zip")
	require.NoError(t, os.WriteFile(path, outer.Bytes(), 0o600))
	return path
}

func TestRecoveryFailsClosed(t *testing.T) {
	t.Run("ambiguous candidates", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(
			cardRecord("card", "Card", 10_000, []any{"missing"}, "", "", ""),
			textRecord("orphan-1", 10_500, "one"),
			textRecord("orphan-2", 10_500, "two"),
		), nil)
		_, err := Convert(input, Options{Strict: true})
		require.ErrorContains(t, err, "ambiguous or missing")
	})

	t.Run("orphan without broken card", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(
			cardRecord("card", "Card", 10_000, []any{"direct"}, "", "", ""),
			textRecord("direct", 10_100, "direct"),
			textRecord("orphan", 20_000, "orphan"),
		), nil)
		_, err := Convert(input, Options{Strict: true})
		require.ErrorContains(t, err, "orphan text blocks remain")
	})
}

func TestPriorityMappingsAndDueWithoutTimezone(t *testing.T) {
	priorities := []struct {
		raw    string
		native int64
	}{
		{"🔥 P0 Срочно + важно", 5},
		{"🧭 P1 Важно, не срочно", 3},
		{"⚡ P2 Срочно, не важно", 2},
		{"🧊 P3 Backlog", 1},
		{"unknown", 0},
	}
	records := []fixtureRecord{}
	for i, priority := range priorities {
		cardID := fmt.Sprintf("card-%d", i)
		textID := fmt.Sprintf("text-%d", i)
		due := ""
		if i == 0 {
			due = "2026-08-04 00:00:00"
		}
		records = append(records,
			cardRecord(cardID, cardID, 10_000+i*2_000, []any{textID}, priority.raw, due, ""),
			textRecord(textID, 10_100+i*2_000, "description"),
		)
	}
	input := writeSyntheticPackage(t, baseBoardRecords(records...), nil)
	result, err := Convert(input, Options{Strict: true})
	require.NoError(t, err)
	for i, task := range result.Normalized.Boards[0].Tasks {
		require.Equal(t, priorities[i].native, task.NativePriority)
	}
	require.Equal(t, "2026-08-04", result.Normalized.Boards[0].Tasks[0].ParsedDueCandidate)
	require.Empty(t, result.Normalized.Boards[0].Tasks[0].NativeDueDate)
}

func TestAssigneeMappingSupportsManyToOneAndExplicitUnassigned(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card-a", "A", 10_000, []any{"text-a"}, "", "", "Alpha"),
		textRecord("text-a", 10_100, "a"),
		cardRecord("card-b", "B", 20_000, []any{"text-b"}, "", "", "Beta"),
		textRecord("text-b", 20_100, "b"),
		cardRecord("card-c", "C", 30_000, []any{"text-c"}, "", "", "Gamma"),
		textRecord("text-c", 30_100, "c"),
	), nil)
	mapping := &AssigneeMap{SchemaVersion: "1", Mappings: []AssigneeMapping{
		{SourceRaw: "Alpha", TargetUsernameOrEmail: "user"},
		{SourceRaw: "Beta", TargetUsernameOrEmail: "user"},
		{SourceRaw: "Gamma", Unassigned: true},
	}}
	result, err := Convert(input, Options{Strict: true, Assignees: mapping})
	require.NoError(t, err)
	tasks := result.Normalized.Boards[0].Tasks
	require.Equal(t, "user", tasks[0].AssigneeTarget)
	require.Equal(t, "user", tasks[1].AssigneeTarget)
	require.True(t, tasks[2].AssigneeUnassigned)
	require.Empty(t, result.Reconciliation.UnknownAssignees)
}

func TestArchiveValidation(t *testing.T) {
	t.Run("SHA mismatch", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(), nil)
		_, err := Analyze(input, Options{ExpectedSHA256: strings.Repeat("0", 64), Strict: true})
		require.ErrorContains(t, err, "SHA-256 mismatch")
	})

	t.Run("path traversal", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(), map[string][]byte{"../escape": []byte("bad")})
		_, err := Analyze(input, Options{Strict: true})
		require.ErrorContains(t, err, "unsafe path")
	})

	t.Run("malformed nested ZIP", func(t *testing.T) {
		input := writeMalformedNestedPackage(t)
		_, err := Analyze(input, Options{Strict: true})
		require.ErrorContains(t, err, "open nested .boardarchive ZIP")
	})

	t.Run("duplicate known JSON key", func(t *testing.T) {
		input := writeRawSyntheticPackage(t, []byte(`{"type":"board","type":"block","data":{}}`+"\n"))
		_, err := Analyze(input, Options{Strict: true})
		require.ErrorContains(t, err, "duplicate key")
	})

	t.Run("malformed JSONL", func(t *testing.T) {
		input := writeRawSyntheticPackage(t, []byte("{not-json\n"))
		_, err := Analyze(input, Options{Strict: true})
		require.ErrorContains(t, err, "decode")
	})
}

func TestStrictSourceSchemaRejectsUnknownFieldsRelationshipsAndAmbiguousProperties(t *testing.T) {
	t.Run("unknown board field", func(t *testing.T) {
		records := baseBoardRecords()
		records[0].Data["unexpected"] = true
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "unknown field")
	})
	t.Run("unsupported board type", func(t *testing.T) {
		records := baseBoardRecords()
		records[0].Data["type"] = "X"
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "unsupported type")
	})
	t.Run("card parent differs from board", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["parentId"] = "other"
		_, err := Analyze(writeSyntheticPackage(t, baseBoardRecords(card), nil), Options{Strict: true})
		require.ErrorContains(t, err, "parentId must equal boardId")
	})
	t.Run("duplicate property id", func(t *testing.T) {
		records := baseBoardRecords()
		props := records[0].Data["cardProperties"].([]any)
		props[1].(map[string]any)["id"] = props[0].(map[string]any)["id"]
		_, err := Convert(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "duplicate property id")
	})
	t.Run("duplicate status option id", func(t *testing.T) {
		records := baseBoardRecords()
		options := records[0].Data["cardProperties"].([]any)[0].(map[string]any)["options"].([]any)
		options[1].(map[string]any)["id"] = options[0].(map[string]any)["id"]
		_, err := Convert(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "duplicate option id")
	})
	t.Run("structured card property value", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["fields"].(map[string]any)["properties"].(map[string]any)["due"] = []any{"2026-08-04"}
		_, err := Convert(writeSyntheticPackage(t, baseBoardRecords(card), nil), Options{Strict: true})
		require.ErrorContains(t, err, "must be a string or null")
	})
	t.Run("unknown card property id", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["fields"].(map[string]any)["properties"].(map[string]any)["unknown"] = "value"
		_, err := Analyze(writeSyntheticPackage(t, baseBoardRecords(card), nil), Options{Strict: true})
		require.ErrorContains(t, err, "unknown property id")
	})
	t.Run("unknown status option id", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["fields"].(map[string]any)["properties"].(map[string]any)["status"] = "not-an-option"
		_, err := Analyze(writeSyntheticPackage(t, baseBoardRecords(card), nil), Options{Strict: true})
		require.ErrorContains(t, err, "unknown option id")
	})
	t.Run("additional property structural value", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["fields"].(map[string]any)["properties"].(map[string]any)["extra"] = []any{"value"}
		records := baseBoardRecords(card)
		props := records[0].Data["cardProperties"].([]any)
		records[0].Data["cardProperties"] = append(props, map[string]any{"id": "extra", "name": "Extra", "type": "text", "options": []any{}})
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "must be a string or null")
	})
	t.Run("additional scalar property is preserved", func(t *testing.T) {
		card := cardRecord("card", "Card", 10_000, []any{}, "", "", "")
		card.Data["fields"].(map[string]any)["properties"].(map[string]any)["extra"] = "value"
		records := baseBoardRecords(card)
		props := records[0].Data["cardProperties"].([]any)
		records[0].Data["cardProperties"] = append(props, map[string]any{"id": "extra", "name": "Extra", "type": "text", "options": []any{}})
		result, err := Convert(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.NoError(t, err)
		require.Equal(t, "value", result.Normalized.Boards[0].Tasks[0].SourceProperties["Extra"])
	})
	t.Run("text property defines options", func(t *testing.T) {
		records := baseBoardRecords()
		props := records[0].Data["cardProperties"].([]any)
		props[1].(map[string]any)["options"] = []any{map[string]any{"id": "x", "value": "X", "color": "default"}}
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "must not define options")
	})

	t.Run("board member references unknown board", func(t *testing.T) {
		records := baseBoardRecords()
		records = append(records, fixtureRecord{Type: "boardMember", Data: map[string]any{"boardId": "missing", "userId": "user", "minimumRole": "viewer", "roles": "", "schemeAdmin": false, "schemeCommenter": false, "schemeEditor": false, "schemeViewer": true, "synthetic": false}})
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "unknown board")
	})
}

func TestDeletedAndTemplateCardsFailClosed(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(fixtureRecord)
	}{
		{"deleted", func(card fixtureRecord) { card.Data["deleteAt"] = 1 }},
		{"template", func(card fixtureRecord) { card.Data["fields"].(map[string]any)["isTemplate"] = true }},
	} {
		t.Run(test.name, func(t *testing.T) {
			card := cardRecord("card", "Card", 10_000, []any{"text"}, "", "", "")
			test.mutate(card)
			input := writeSyntheticPackage(t, baseBoardRecords(card, textRecord("text", 10_100, "description")), nil)
			_, err := Convert(input, Options{Strict: true})
			require.ErrorContains(t, err, "deleted/template cards")
		})
	}
}

func TestArtifactsAreDeterministic(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card", "Card", 10_000, []any{"text"}, "🔥 P0 Срочно + важно", "2026-08-04", "Alpha"),
		textRecord("text", 10_100, "description"),
	), nil)
	opts := Options{Strict: true, Timezone: "UTC", VikunjaVersion: "2.5.0", RepoSHA: strings.Repeat("a", 40)}
	first, err := Convert(input, opts)
	require.NoError(t, err)
	second, err := Convert(input, opts)
	require.NoError(t, err)
	require.Equal(t, sha256Hex(first.NormalizedJSON), sha256Hex(second.NormalizedJSON))
	require.Equal(t, sha256Hex(first.VikunjaZip), sha256Hex(second.VikunjaZip))
	require.Equal(t, sha256Hex(first.ReconciliationJSON), sha256Hex(second.ReconciliationJSON))
	require.Equal(t, sha256Hex(first.ReconciliationCSV), sha256Hex(second.ReconciliationCSV))
	require.Equal(t, sha256Hex(first.AssigneeTemplate), sha256Hex(second.AssigneeTemplate))
	require.Equal(t, sha256Hex(first.ManifestJSON), sha256Hex(second.ManifestJSON))
	require.Equal(t, sha256Hex(first.README), sha256Hex(second.README))
	zr, err := zip.NewReader(bytes.NewReader(first.VikunjaZip), int64(len(first.VikunjaZip)))
	require.NoError(t, err)
	zr2, err := zip.NewReader(bytes.NewReader(second.VikunjaZip), int64(len(second.VikunjaZip)))
	require.NoError(t, err)
	require.Len(t, zr.File, 3)
	require.Len(t, zr2.File, 3)
	for i, entry := range zr.File {
		require.True(t, entry.Modified.Equal(deterministicZipTime))
		require.Equal(t, os.FileMode(0o600), entry.Mode().Perm())
		require.Equal(t, zr2.File[i].Name, entry.Name)
		require.Equal(t, zr2.File[i].Extra, entry.Extra)
		require.Equal(t, zr2.File[i].CreatorVersion, entry.CreatorVersion)
		require.Equal(t, zr2.File[i].ReaderVersion, entry.ReaderVersion)
		require.Empty(t, entry.Comment)
	}
	require.Equal(t, ConverterVersion, first.Manifest.ToolVersion)
	require.Len(t, first.Manifest.ConfigSHA256, 64)
	require.NotEmpty(t, first.Manifest.GeneratedAt)
	require.False(t, first.Manifest.RepoDirty)
	require.Contains(t, string(first.README), "/api/v2/migration/vikunja-file/migrate")
	require.NotContains(t, string(first.README), "/api/v1/migration/vikunja-file/migrate")
}

func TestVerifyRejectsDuplicateSourceIDs(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card", "Card", 10_000, []any{"text"}, "", "", ""),
		textRecord("text", 10_100, "description"),
	), nil)
	result, err := Convert(input, Options{Strict: true})
	require.NoError(t, err)
	var normalized NormalizedArchive
	require.NoError(t, json.Unmarshal(result.NormalizedJSON, &normalized))
	normalized.Boards[0].Tasks = append(normalized.Boards[0].Tasks, normalized.Boards[0].Tasks[0])
	normalized.Counts.Cards++
	duplicateJSON, err := marshalPretty(&normalized)
	require.NoError(t, err)
	_, err = Verify(input, duplicateJSON, result.VikunjaZip, Options{Strict: true})
	require.ErrorContains(t, err, "duplicate normalized source card id")
}

func TestBacklogBucketsFollowStatusSchema(t *testing.T) {
	card := cardRecord("card", "Card", 10_000, []any{"text"}, "", "", "")
	card.Data["fields"].(map[string]any)["properties"].(map[string]any)["status"] = "new"
	records := baseBoardRecords(card, textRecord("text", 10_100, "description"))
	boardProperties := records[0].Data["cardProperties"].([]any)
	statusProperty := boardProperties[0].(map[string]any)
	statusProperty["options"] = []any{
		map[string]any{"id": "new", "value": "Новая", "color": "default"},
		map[string]any{"id": "working", "value": "В работе", "color": "default"},
		map[string]any{"id": "sorted", "value": "Разобрана", "color": "default"},
	}
	input := writeSyntheticPackage(t, records, nil)
	result, err := Convert(input, Options{Strict: true})
	require.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(result.VikunjaZip), int64(len(result.VikunjaZip)))
	require.NoError(t, err)
	var data []byte
	for _, f := range zr.File {
		if f.Name == "data.json" {
			r, openErr := f.Open()
			require.NoError(t, openErr)
			data, err = io.ReadAll(r)
			require.NoError(t, err)
			require.NoError(t, r.Close())
		}
	}
	var projects []*models.ProjectWithTasksAndBuckets
	require.NoError(t, json.Unmarshal(data, &projects))
	require.Len(t, projects, 2)
	bucketNames := []string{}
	for _, bucket := range projects[1].Buckets {
		bucketNames = append(bucketNames, bucket.Title)
	}
	require.Equal(t, []string{"Новая", "В работе", "Разобрана"}, bucketNames)
}

func TestDirectDescriptionLinksFailClosed(t *testing.T) {
	t.Run("same text referenced by two cards", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(
			cardRecord("card-a", "A", 10_000, []any{"text"}, "", "", ""),
			cardRecord("card-b", "B", 20_000, []any{"text"}, "", "", ""),
			textRecord("text", 10_100, "description"),
		), nil)
		_, err := Convert(input, Options{Strict: true})
		require.ErrorContains(t, err, "direct description link is not one-to-one")
	})

	t.Run("multiple content refs", func(t *testing.T) {
		input := writeSyntheticPackage(t, baseBoardRecords(
			cardRecord("card", "Card", 10_000, []any{"text-a", "text-b"}, "", "", ""),
			textRecord("text-a", 10_100, "a"),
			textRecord("text-b", 10_200, "b"),
		), nil)
		_, err := Convert(input, Options{Strict: true})
		require.ErrorContains(t, err, "unsupported contentOrder length")
	})
}

func TestReconciliationUsesGlobalTargetTaskIDs(t *testing.T) {
	statusProperty := rawCardProperty{ID: "status", Name: "Status", Type: "select", Options: []rawPropertyOption{{ID: "todo", Value: "Todo"}, {ID: "doing", Value: "In Progress"}, {ID: "done", Value: "Done"}}}
	requiredProperties := []rawCardProperty{statusProperty, {ID: "assignee", Name: "Кто делает", Type: "text"}, {ID: "priority", Name: "Приоритет", Type: "text"}, {ID: "due", Name: "Срок", Type: "text"}}
	parsed := &parsedArchive{
		SourceHash: strings.Repeat("a", 64),
		NestedHash: strings.Repeat("b", 64),
		Boards: []*rawBoard{
			{ID: "board-a", Title: "A", CardProperties: requiredProperties},
			{ID: "board-b", Title: "B", CardProperties: requiredProperties},
		},
		Cards: []*rawBlock{
			{ID: "card-a", BoardID: "board-a", Title: "A", CreateAt: 1, Fields: rawBlockFields{ContentOrder: []string{"text-a"}, Properties: map[string]any{"status": "todo"}}},
			{ID: "card-b", BoardID: "board-b", Title: "B", CreateAt: 2, Fields: rawBlockFields{ContentOrder: []string{"text-b"}, Properties: map[string]any{"status": "todo"}}},
		},
		Texts: []*rawBlock{
			{ID: "text-a", BoardID: "board-a", Title: "A", CreateAt: 2},
			{ID: "text-b", BoardID: "board-b", Title: "B", CreateAt: 3},
		},
	}
	normalized, reconciliation, err := normalizeParsed(parsed, Options{Strict: true, VikunjaVersion: DefaultVikunjaVersion})
	require.NoError(t, err)
	require.Len(t, normalized.Boards, 2)
	require.Equal(t, int64(1), reconciliation.CardMappings[0].TargetTaskLegacyID)
	require.Equal(t, int64(2), reconciliation.CardMappings[1].TargetTaskLegacyID)
	structure, err := buildVikunjaStructure(normalized)
	require.NoError(t, err)
	require.Equal(t, int64(1), structure[1].Tasks[0].ID)
	require.Equal(t, int64(2), structure[2].Tasks[0].ID)
}

func TestSourceAndTargetVersionsFailClosed(t *testing.T) {
	records := baseBoardRecords()
	cases := []struct {
		name     string
		data     []byte
		include  bool
		contains string
	}{
		{"missing", nil, false, "exactly one valid version.json"},
		{"malformed", []byte(`{"version":`), true, "decode version.json"},
		{"unsupported", []byte(`{"version":9,"date":1722686400000}`), true, "unsupported Focalboard archive version"},
		{"unknown field", []byte(`{"version":2,"date":1722686400000,"extra":true}`), true, "unknown field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := writeSyntheticPackageWithVersion(t, records, nil, tc.data, tc.include)
			_, err := Analyze(input, Options{Strict: true})
			require.ErrorContains(t, err, tc.contains)
		})
	}
	input := writeSyntheticPackage(t, records, nil)
	_, err := Convert(input, Options{Strict: true, VikunjaVersion: "2.6.0"})
	require.ErrorContains(t, err, "pinned")
}

func TestViewsFailClosed(t *testing.T) {
	t.Run("non kanban", func(t *testing.T) {
		records := baseBoardRecords()
		records[1].Data["fields"] = map[string]any{"viewType": "table"}
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "unsupported Focalboard view type")
	})
	t.Run("unknown board", func(t *testing.T) {
		records := baseBoardRecords()
		records[1].Data["boardId"] = "missing-board"
		_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
		require.ErrorContains(t, err, "references unknown board")
	})
}

func TestUnequalRecoveryCandidatesCannotBeSilentlyDropped(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card", "Card", 10_000, []any{"missing"}, "", "", ""),
		textRecord("orphan-near", 10_600, "near"),
		textRecord("orphan-far", 10_800, "far"),
	), nil)
	_, err := Convert(input, Options{Strict: true})
	require.ErrorContains(t, err, "orphan text blocks remain")
}

func TestIncompleteAssigneeTemplateRemainsUnmapped(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(
		cardRecord("card", "Card", 10_000, []any{"text"}, "", "", "Alpha"),
		textRecord("text", 10_100, "description"),
	), nil)
	first, err := Convert(input, Options{Strict: true})
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "assignees.yaml")
	require.NoError(t, os.WriteFile(path, first.AssigneeTemplate, 0o600))
	mapping, err := LoadAssigneeMap(path)
	require.NoError(t, err)
	second, err := Convert(input, Options{Strict: true, Assignees: mapping})
	require.NoError(t, err)
	require.Equal(t, []string{"Alpha"}, second.Reconciliation.UnknownAssignees)
	require.False(t, second.Reconciliation.CardMappings[0].AssigneeMapped)
	codes := []string{}
	for _, warning := range second.Reconciliation.Warnings {
		codes = append(codes, warning.Code)
	}
	require.Contains(t, codes, "unmapped_assignee")
}

func TestVerifyRejectsExcessiveVikunjaZipEntries(t *testing.T) {
	input := writeSyntheticPackage(t, baseBoardRecords(), nil)
	result, err := Convert(input, Options{Strict: true})
	require.NoError(t, err)
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for i := 0; i < 17; i++ {
		entry, createErr := writer.Create(fmt.Sprintf("entry-%02d", i))
		require.NoError(t, createErr)
		_, createErr = entry.Write([]byte("x"))
		require.NoError(t, createErr)
	}
	require.NoError(t, writer.Close())
	err = verifyNormalizedAndZip(result.Normalized, buffer.Bytes(), result.Reconciliation)
	require.ErrorContains(t, err, "too many entries")
}

func TestRequiredPropertySchemaFailsClosed(t *testing.T) {
	records := baseBoardRecords()
	properties := records[0].Data["cardProperties"].([]any)
	records[0].Data["cardProperties"] = properties[:len(properties)-1]
	_, err := Analyze(writeSyntheticPackage(t, records, nil), Options{Strict: true})
	require.ErrorContains(t, err, "missing required property")
}

func TestCoreMigrationAPIsRequireStrictMode(t *testing.T) {
	_, err := Analyze("not-read", Options{})
	require.ErrorIs(t, err, errStrictModeRequired)
	_, err = Convert("not-read", Options{})
	require.ErrorIs(t, err, errStrictModeRequired)
	_, err = Verify("not-read", nil, nil, Options{})
	require.ErrorIs(t, err, errStrictModeRequired)
}
