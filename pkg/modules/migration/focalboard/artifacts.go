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
	"html"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/yuin/goldmark"
)

var deterministicZipTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type sourceMarker struct {
	SourceSystem        string `json:"source_system"`
	SourceArchiveSHA256 string `json:"source_archive_sha256"`
	SourceBoardID       string `json:"source_board_id"`
	SourceCardID        string `json:"source_card_id"`
	SourceTextID        string `json:"source_text_id,omitempty"`
	RunID               string `json:"run_id"`
}

func marshalPretty(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func renderDescription(task NormalizedTask, runID string) (string, error) {
	var rendered bytes.Buffer
	if task.Description != "" {
		if err := goldmark.Convert([]byte(task.Description), &rendered); err != nil {
			return "", fmt.Errorf("render description for card %s: %w", task.SourceCardID, err)
		}
	}
	markerBytes, err := json.Marshal(sourceMarker{
		SourceSystem:        "focalboard",
		SourceArchiveSHA256: task.SourceArchiveSHA256,
		SourceBoardID:       task.SourceBoardID,
		SourceCardID:        task.SourceCardID,
		SourceTextID:        task.SourceTextID,
		RunID:               runID,
	})
	if err != nil {
		return "", err
	}
	metadataBytes, err := json.Marshal(struct {
		SourceBoardID         string                `json:"source_board_id"`
		SourceCardID          string                `json:"source_card_id"`
		SourceTextID          string                `json:"source_text_id,omitempty"`
		DescriptionLinkMethod DescriptionLinkMethod `json:"description_link_method"`
		StatusRaw             string                `json:"status_raw"`
		SourceProperties      map[string]string     `json:"source_properties"`
		AssigneeRaw           string                `json:"assignee_raw"`
		PriorityRaw           string                `json:"priority_raw"`
		DueRaw                string                `json:"due_raw"`
		SourceCreatedAt       string                `json:"source_created_at"`
		SourceUpdatedAt       string                `json:"source_updated_at"`
	}{task.SourceBoardID, task.SourceCardID, task.SourceTextID, task.DescriptionLinkMethod, task.StatusRaw, task.SourceProperties, task.AssigneeRaw, task.PriorityRaw, task.DueRaw, task.SourceCreatedAt, task.SourceUpdatedAt})
	if err != nil {
		return "", err
	}
	if rendered.Len() > 0 && !strings.HasSuffix(rendered.String(), "\n") {
		rendered.WriteByte('\n')
	}
	rendered.WriteString("<!-- focalboard-migration:")
	rendered.Write(markerBytes)
	rendered.WriteString(" -->\n<details><summary>Focalboard migration metadata</summary><pre>")
	rendered.WriteString(html.EscapeString(string(metadataBytes)))
	rendered.WriteString("</pre></details>")
	return rendered.String(), nil
}

func bucketTitles(board NormalizedBoard) ([]string, error) {
	joined := strings.Join(board.StatusOptions, "\x00")
	switch joined {
	case strings.Join([]string{"Done", "In Progress", "Todo"}, "\x00"):
		return []string{"Todo", "In Progress", "Done"}, nil
	case strings.Join([]string{"В работе", "Новая", "Разобрана"}, "\x00"):
		return []string{"Новая", "В работе", "Разобрана"}, nil
	default:
		return nil, fmt.Errorf("unsupported normalized Status option schema for board %s", board.SourceBoardID)
	}
}

func buildVikunjaStructure(normalized *NormalizedArchive) ([]*models.ProjectWithTasksAndBuckets, error) {
	rootID := int64(1)
	projects := []*models.ProjectWithTasksAndBuckets{{
		Project: models.Project{ID: rootID, Title: RootProjectTitle, Position: 100},
	}}
	globalTaskID := int64(1)
	for boardIndex, board := range normalized.Boards {
		projectID := int64(boardIndex + 2)
		viewID := int64(10_000 + boardIndex + 1)
		titles, err := bucketTitles(board)
		if err != nil {
			return nil, err
		}
		buckets := make([]*models.Bucket, 0, len(titles))
		bucketByStatus := map[string]int64{}
		for bucketIndex, title := range titles {
			bucketID := int64(100_000 + (boardIndex+1)*10 + bucketIndex + 1)
			buckets = append(buckets, &models.Bucket{
				ID:            bucketID,
				Title:         title,
				ProjectViewID: viewID,
				Position:      float64((bucketIndex + 1) * 100),
			})
			bucketByStatus[title] = bucketID
		}
		view := &models.ProjectView{
			ID:                      viewID,
			ProjectID:               projectID,
			Title:                   "Kanban",
			ViewKind:                models.ProjectViewKindKanban,
			Position:                100,
			BucketConfigurationMode: models.BucketConfigurationModeManual,
			DefaultBucketID:         buckets[0].ID,
			DoneBucketID:            buckets[len(buckets)-1].ID,
		}
		project := &models.ProjectWithTasksAndBuckets{
			Project: models.Project{
				ID:              projectID,
				Title:           board.Title,
				ParentProjectID: &rootID,
				Position:        float64((boardIndex + 1) * 100),
				Views:           []*models.ProjectView{view},
			},
			Buckets: buckets,
			Tasks:   []*models.TaskWithComments{},
		}
		for taskIndex, task := range board.Tasks {
			bucketID, exists := bucketByStatus[task.StatusRaw]
			if !exists {
				return nil, fmt.Errorf("card %s has status %q which is not a bucket on its board", task.SourceCardID, task.StatusRaw)
			}
			description, err := renderDescription(task, normalized.MigrationRun.RunID)
			if err != nil {
				return nil, err
			}
			vikunjaTask := models.Task{
				ID:          globalTaskID,
				Title:       task.Title,
				Description: description,
				Priority:    task.NativePriority,
				BucketID:    bucketID,
				Position:    float64((taskIndex + 1) * 65536),
				// Leave Done false in the serialized structure. The current file
				// importer first creates the task in the default bucket and then
				// moves it to BucketID; that transition sets Done and DoneAt. If
				// Done is pre-set here, the transition sees no state change and
				// persists the task as not done.
				Done: false,
			}
			if task.NativeDueDate != "" {
				due, err := time.Parse(time.RFC3339, task.NativeDueDate)
				if err != nil {
					return nil, fmt.Errorf("parse normalized due date for card %s: %w", task.SourceCardID, err)
				}
				vikunjaTask.DueDate = due
			}
			project.Tasks = append(project.Tasks, &models.TaskWithComments{Task: vikunjaTask, Comments: []*models.TaskComment{}})
			globalTaskID++
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func writeDeterministicZipEntry(zw *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.Modified = deterministicZipTime
	header.SetMode(0o600)
	header.CreatorVersion = 3 << 8
	header.ReaderVersion = 20
	header.Extra = nil
	header.Comment = ""
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func buildVikunjaZip(normalized *NormalizedArchive) ([]byte, error) {
	structure, err := buildVikunjaStructure(normalized)
	if err != nil {
		return nil, err
	}
	dataJSON, err := json.Marshal(structure)
	if err != nil {
		return nil, fmt.Errorf("marshal Vikunja data.json: %w", err)
	}
	var output bytes.Buffer
	zw := zip.NewWriter(&output)
	if err := writeDeterministicZipEntry(zw, "VERSION", []byte(normalized.MigrationRun.VikunjaVersion)); err != nil {
		return nil, err
	}
	if err := writeDeterministicZipEntry(zw, "data.json", dataJSON); err != nil {
		return nil, err
	}
	if err := writeDeterministicZipEntry(zw, "filters.json", []byte("[]")); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func buildReconciliationCSV(reconciliation *Reconciliation) ([]byte, error) {
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{
		"source_board_id", "source_card_id", "source_text_id", "source_created_at", "source_updated_at", "target_project_legacy_id", "target_task_legacy_id",
		"description_link_method", "native_priority", "native_due_date_set", "assignee_mapped",
	}); err != nil {
		return nil, err
	}
	for _, row := range reconciliation.CardMappings {
		if err := writer.Write([]string{
			row.SourceBoardID,
			row.SourceCardID,
			row.SourceTextID,
			row.SourceCreatedAt,
			row.SourceUpdatedAt,
			fmt.Sprint(row.TargetProjectLegacyID),
			fmt.Sprint(row.TargetTaskLegacyID),
			string(row.DescriptionLinkMethod),
			fmt.Sprint(row.NativePriority),
			fmt.Sprint(row.NativeDueDateSet),
			fmt.Sprint(row.AssigneeMapped),
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func buildAssigneeTemplate(normalized *NormalizedArchive) ([]byte, error) {
	values := map[string]struct{}{}
	for _, board := range normalized.Boards {
		for _, task := range board.Tasks {
			if task.AssigneeRaw != "" {
				values[task.AssigneeRaw] = struct{}{}
			}
		}
	}
	sorted := make([]string, 0, len(values))
	for value := range values {
		sorted = append(sorted, value)
	}
	sort.Strings(sorted)
	var output strings.Builder
	output.WriteString("schema_version: \"1\"\nmappings:\n")
	for _, value := range sorted {
		quoted, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal assignee source value: %w", err)
		}
		output.WriteString("  - source_raw: ")
		output.Write(quoted)
		output.WriteString("\n    target_username_or_email: \"\"\n    unassigned: false\n")
	}
	return []byte(output.String()), nil
}

func buildREADME(normalized *NormalizedArchive) []byte {
	dueNote := "Native due dates are omitted because --timezone was not set."
	if normalized.MigrationRun.Timezone != "" {
		dueNote = "Strict ISO due dates use timezone " + normalized.MigrationRun.Timezone + "."
	}
	return []byte(fmt.Sprintf(`# Focalboard → Vikunja migration artifacts

Run ID: %s
Source SHA-256: %s
Vikunja file format version: %s

## Import into a clean Vikunja workspace

The native file importer is not an upsert. Re-import only into a new or cleaned workspace.

UI: Settings → Import from other services → Vikunja file → select vikunja-import.zip.

API equivalent after the instance exists:

    test -n "$VIKUNJA_URL" && test -n "$VIKUNJA_TOKEN"
    curl --fail-with-body --silent --show-error \
      -H "Authorization: Bearer $VIKUNJA_TOKEN" \
      -F "import=@vikunja-import.zip;type=application/zip" \
      "$VIKUNJA_URL/api/v2/migration/vikunja-file/migrate"

## Constraints

- %s
- Free-text due values are preserved only as source metadata and are never guessed.
- Tasks remain unassigned in the native ZIP. Fill assignees-map.template.yaml after users exist, then apply assignments through the Vikunja API and reconcile by source card ID.
- Source IDs and the run ID are embedded in every imported task description and in reconciliation.json/csv.
- Live upload, API reconciliation, assignment application, backup/restore, and cutover freeze were not run by this converter.
`, normalized.MigrationRun.RunID, normalized.SourceArchiveSHA256, normalized.MigrationRun.VikunjaVersion, dueNote))
}

func Convert(inputPath string, opts Options) (*Result, error) {
	if !opts.Strict {
		return nil, errStrictModeRequired
	}
	if opts.VikunjaVersion == "" {
		opts.VikunjaVersion = DefaultVikunjaVersion
	}
	parsed, err := parsePackage(inputPath, opts)
	if err != nil {
		return nil, err
	}
	normalized, reconciliation, err := normalizeParsed(parsed, opts)
	if err != nil {
		return nil, err
	}
	normalizedJSON, err := marshalPretty(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal normalized JSON: %w", err)
	}
	vikunjaZip, err := buildVikunjaZip(normalized)
	if err != nil {
		return nil, err
	}
	if err := verifyNormalizedAndZip(normalized, vikunjaZip, reconciliation); err != nil {
		return nil, fmt.Errorf("verify generated artifacts: %w", err)
	}
	reconciliation.Verified = true
	reconciliationJSON, err := marshalPretty(reconciliation)
	if err != nil {
		return nil, fmt.Errorf("marshal reconciliation JSON: %w", err)
	}
	reconciliationCSV, err := buildReconciliationCSV(reconciliation)
	if err != nil {
		return nil, fmt.Errorf("marshal reconciliation CSV: %w", err)
	}
	assigneeTemplate, err := buildAssigneeTemplate(normalized)
	if err != nil {
		return nil, err
	}
	readme := buildREADME(normalized)
	configHash, err := configurationHash(opts)
	if err != nil {
		return nil, err
	}
	generatedAt := parsed.MaxUpdated.Format(time.RFC3339Nano)
	manifest := &RunManifest{
		SchemaVersion:       NormalizedSchemaVersion,
		RunID:               normalized.MigrationRun.RunID,
		RepoSHA:             opts.RepoSHA,
		RepoDirty:           opts.RepoDirty,
		ToolVersion:         ConverterVersion,
		GeneratedAt:         generatedAt,
		ConfigSHA256:        configHash,
		SourceArchiveSHA256: parsed.SourceHash,
		SourceNestedSHA256:  parsed.NestedHash,
		SourceVersion:       parsed.SourceVersion,
		SourceMaxUpdatedAt:  parsed.MaxUpdated.Format(time.RFC3339Nano),
		Config: ManifestConfig{
			Strict:         opts.Strict,
			Timezone:       opts.Timezone,
			VikunjaVersion: opts.VikunjaVersion,
		},
		Counts: normalized.Counts,
		ArtifactSHA256: map[string]string{
			"README.md":                   sha256Hex(readme),
			"assignees-map.template.yaml": sha256Hex(assigneeTemplate),
			"normalized-focalboard.json":  sha256Hex(normalizedJSON),
			"reconciliation.csv":          sha256Hex(reconciliationCSV),
			"reconciliation.json":         sha256Hex(reconciliationJSON),
			"vikunja-import.zip":          sha256Hex(vikunjaZip),
		},
	}
	manifestJSON, err := marshalPretty(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal run manifest: %w", err)
	}
	return &Result{
		Normalized:         normalized,
		Reconciliation:     reconciliation,
		Manifest:           manifest,
		NormalizedJSON:     normalizedJSON,
		VikunjaZip:         vikunjaZip,
		ReconciliationJSON: reconciliationJSON,
		ReconciliationCSV:  reconciliationCSV,
		AssigneeTemplate:   assigneeTemplate,
		ManifestJSON:       manifestJSON,
		README:             readme,
	}, nil
}
