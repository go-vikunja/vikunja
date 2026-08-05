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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

func extractMarker(description string) (sourceMarker, error) {
	const prefix = "<!-- focalboard-migration:"
	const suffix = " -->"
	start := strings.LastIndex(description, prefix)
	if start < 0 {
		return sourceMarker{}, fmt.Errorf("missing focalboard source marker")
	}
	start += len(prefix)
	endRelative := strings.Index(description[start:], suffix)
	if endRelative < 0 {
		return sourceMarker{}, fmt.Errorf("unterminated focalboard source marker")
	}
	var marker sourceMarker
	if err := json.Unmarshal([]byte(description[start:start+endRelative]), &marker); err != nil {
		return sourceMarker{}, fmt.Errorf("decode focalboard source marker: %w", err)
	}
	if marker.SourceSystem != "focalboard" || marker.SourceArchiveSHA256 == "" || marker.SourceBoardID == "" || marker.SourceCardID == "" || marker.RunID == "" {
		return sourceMarker{}, fmt.Errorf("incomplete focalboard source marker")
	}
	return marker, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	if file.UncompressedSize64 > maxNestedEntrySize {
		return nil, fmt.Errorf("zip entry %q exceeds size limit", file.Name)
	}
	r, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return readLimited(r, maxNestedEntrySize)
}

func verifyNormalizedAndZip(normalized *NormalizedArchive, vikunjaZip []byte, reconciliation *Reconciliation) error {
	if normalized.SchemaVersion != NormalizedSchemaVersion {
		return fmt.Errorf("unsupported normalized schema version %q", normalized.SchemaVersion)
	}
	if normalized.SourceSystem != "focalboard" || normalized.SourceArchiveSHA256 == "" || normalized.MigrationRun.RunID == "" {
		return fmt.Errorf("normalized archive identity is incomplete")
	}
	if normalized.SourceVersion != SupportedFocalboardVersion {
		return fmt.Errorf("unsupported normalized Focalboard source version %d", normalized.SourceVersion)
	}
	if normalized.MigrationRun.VikunjaVersion != DefaultVikunjaVersion {
		return fmt.Errorf("unsupported normalized Vikunja VERSION %q", normalized.MigrationRun.VikunjaVersion)
	}
	normalizedIDs := map[string]NormalizedTask{}
	normalizedTextIDs := map[string]struct{}{}
	boardIDs := map[string]struct{}{}
	viewIDs := map[string]struct{}{}
	normalizedTaskCount := 0
	for _, board := range normalized.Boards {
		if board.SourceBoardID == "" {
			return fmt.Errorf("normalized board has empty source id")
		}
		if _, exists := boardIDs[board.SourceBoardID]; exists {
			return fmt.Errorf("duplicate normalized source board id %s", board.SourceBoardID)
		}
		boardIDs[board.SourceBoardID] = struct{}{}
		if _, err := bucketTitles(board); err != nil {
			return err
		}
		for _, viewID := range board.SourceViewIDs {
			if viewID == "" {
				return fmt.Errorf("normalized board has empty source view id")
			}
			if _, exists := viewIDs[viewID]; exists {
				return fmt.Errorf("duplicate normalized source view id %s", viewID)
			}
			viewIDs[viewID] = struct{}{}
		}
		for _, task := range board.Tasks {
			if task.SourceBoardID != board.SourceBoardID || task.SourceCardID == "" {
				return fmt.Errorf("normalized task source identity is inconsistent")
			}
			key := task.SourceBoardID + "\x00" + task.SourceCardID
			if _, exists := normalizedIDs[key]; exists {
				return fmt.Errorf("duplicate normalized source card id %s/%s", task.SourceBoardID, task.SourceCardID)
			}
			switch task.DescriptionLinkMethod {
			case DescriptionNone:
				if task.SourceTextID != "" {
					return fmt.Errorf("normalized card %s/%s has source text id with no description", task.SourceBoardID, task.SourceCardID)
				}
			case DescriptionDirect, DescriptionRecovered:
				if task.SourceTextID == "" {
					return fmt.Errorf("normalized card %s/%s is missing source text id", task.SourceBoardID, task.SourceCardID)
				}
				if _, exists := normalizedTextIDs[task.SourceTextID]; exists {
					return fmt.Errorf("duplicate normalized source text id %s", task.SourceTextID)
				}
				normalizedTextIDs[task.SourceTextID] = struct{}{}
			default:
				return fmt.Errorf("normalized card %s/%s has unknown description link method %q", task.SourceBoardID, task.SourceCardID, task.DescriptionLinkMethod)
			}
			normalizedIDs[key] = task
			normalizedTaskCount++
		}
	}
	if normalizedTaskCount != normalized.Counts.Cards {
		return fmt.Errorf("normalized task count mismatch: tasks=%d counts.cards=%d", normalizedTaskCount, normalized.Counts.Cards)
	}
	if len(normalizedTextIDs) != normalized.Counts.FinalDescriptions || len(normalizedTextIDs) != normalized.Counts.TextBlocks {
		return fmt.Errorf("normalized source text count mismatch: ids=%d final_descriptions=%d text_blocks=%d", len(normalizedTextIDs), normalized.Counts.FinalDescriptions, normalized.Counts.TextBlocks)
	}
	if len(viewIDs) != normalized.Counts.KanbanViews {
		return fmt.Errorf("normalized source view count mismatch: ids=%d views=%d", len(viewIDs), normalized.Counts.KanbanViews)
	}

	zr, err := zip.NewReader(bytes.NewReader(vikunjaZip), int64(len(vikunjaZip)))
	if err != nil {
		return fmt.Errorf("open Vikunja ZIP: %w", err)
	}
	if len(zr.File) > 16 {
		return fmt.Errorf("vikunja ZIP has too many entries: %d", len(zr.File))
	}
	entries := map[string]*zip.File{}
	var totalUncompressed uint64
	for _, file := range zr.File {
		totalUncompressed += file.UncompressedSize64
		if totalUncompressed > uint64(MaxVikunjaZipSize) {
			return fmt.Errorf("vikunja ZIP exceeds total uncompressed size limit")
		}
		if err := validateZipEntryName(file.Name); err != nil {
			return err
		}
		if _, exists := entries[file.Name]; exists {
			return fmt.Errorf("duplicate ZIP entry %q", file.Name)
		}
		entries[file.Name] = file
	}
	for _, required := range []string{"VERSION", "data.json", "filters.json"} {
		if entries[required] == nil {
			return fmt.Errorf("vikunja ZIP is missing %s", required)
		}
	}
	version, err := readZipFile(entries["VERSION"])
	if err != nil {
		return err
	}
	if string(version) != normalized.MigrationRun.VikunjaVersion {
		return fmt.Errorf("vikunja ZIP VERSION mismatch: expected %s got %s", normalized.MigrationRun.VikunjaVersion, string(version))
	}
	filters, err := readZipFile(entries["filters.json"])
	if err != nil {
		return err
	}
	var filterData []any
	if err := json.Unmarshal(filters, &filterData); err != nil {
		return fmt.Errorf("decode filters.json: %w", err)
	}
	data, err := readZipFile(entries["data.json"])
	if err != nil {
		return err
	}
	var projects []*models.ProjectWithTasksAndBuckets
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("decode Vikunja data.json: %w", err)
	}
	if len(projects) != len(normalized.Boards)+1 {
		return fmt.Errorf("vikunja project count mismatch: got %d expected %d", len(projects), len(normalized.Boards)+1)
	}
	zipIDs := map[string]struct{}{}
	zipTargetIDs := map[string][2]int64{}
	zipTaskCount := 0
	for projectIndex, project := range projects {
		if projectIndex > 0 {
			expectedBoard := normalized.Boards[projectIndex-1]
			if project.ID != int64(projectIndex+1) || project.Title != expectedBoard.Title {
				return fmt.Errorf("target project legacy id/title mismatch at index %d", projectIndex)
			}
		}
		for _, task := range project.Tasks {
			marker, err := extractMarker(task.Description)
			if err != nil {
				return fmt.Errorf("target task %d: %w", task.ID, err)
			}
			if marker.SourceArchiveSHA256 != normalized.SourceArchiveSHA256 || marker.RunID != normalized.MigrationRun.RunID {
				return fmt.Errorf("target task %d marker archive/run identity mismatch", task.ID)
			}
			key := marker.SourceBoardID + "\x00" + marker.SourceCardID
			if _, exists := zipIDs[key]; exists {
				return fmt.Errorf("duplicate source card id in Vikunja ZIP: %s/%s", marker.SourceBoardID, marker.SourceCardID)
			}
			normalizedTask, exists := normalizedIDs[key]
			if !exists {
				return fmt.Errorf("vikunja ZIP contains source card missing from normalized JSON: %s/%s", marker.SourceBoardID, marker.SourceCardID)
			}
			if marker.SourceTextID != normalizedTask.SourceTextID {
				return fmt.Errorf("target task %d marker source text id mismatch", task.ID)
			}
			zipIDs[key] = struct{}{}
			zipTargetIDs[key] = [2]int64{project.ID, task.ID}
			zipTaskCount++
		}
	}
	if zipTaskCount != normalizedTaskCount {
		return fmt.Errorf("vikunja task count mismatch: zip=%d normalized=%d", zipTaskCount, normalizedTaskCount)
	}
	for key := range normalizedIDs {
		if _, exists := zipIDs[key]; !exists {
			return fmt.Errorf("normalized source card missing from Vikunja ZIP")
		}
	}
	if reconciliation != nil {
		if len(reconciliation.CardMappings) != normalizedTaskCount {
			return fmt.Errorf("reconciliation mapping count mismatch")
		}
		seenMappings := map[string]struct{}{}
		for _, mapping := range reconciliation.CardMappings {
			key := mapping.SourceBoardID + "\x00" + mapping.SourceCardID
			if _, exists := seenMappings[key]; exists {
				return fmt.Errorf("duplicate reconciliation mapping")
			}
			seenMappings[key] = struct{}{}
			target, exists := zipTargetIDs[key]
			if !exists || target[0] != mapping.TargetProjectLegacyID || target[1] != mapping.TargetTaskLegacyID {
				return fmt.Errorf("reconciliation target legacy id mismatch for source card")
			}
		}
	}
	return nil
}

func Verify(inputPath string, normalizedJSON, vikunjaZip []byte, opts Options) (*Reconciliation, error) {
	if !opts.Strict {
		return nil, errStrictModeRequired
	}
	var provided NormalizedArchive
	decoder := json.NewDecoder(bytes.NewReader(normalizedJSON))
	if err := decoder.Decode(&provided); err != nil {
		return nil, fmt.Errorf("decode normalized JSON: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("normalized JSON contains trailing data")
		}
		return nil, fmt.Errorf("decode normalized JSON trailing data: %w", err)
	}
	if opts.VikunjaVersion == "" {
		opts.VikunjaVersion = DefaultVikunjaVersion
	}
	parsed, err := parsePackage(inputPath, opts)
	if err != nil {
		return nil, err
	}
	expected, reconciliation, err := normalizeParsed(parsed, opts)
	if err != nil {
		return nil, err
	}
	if err := verifyNormalizedAndZip(&provided, vikunjaZip, reconciliation); err != nil {
		return nil, err
	}
	expectedJSON, err := marshalPretty(expected)
	if err != nil {
		return nil, err
	}
	if sha256Hex(expectedJSON) != sha256Hex(normalizedJSON) {
		return nil, fmt.Errorf("normalized JSON is not the deterministic artifact for this input/config: expected_sha256=%s actual_sha256=%s", sha256Hex(expectedJSON), sha256Hex(normalizedJSON))
	}
	expectedZip, err := buildVikunjaZip(expected)
	if err != nil {
		return nil, err
	}
	if sha256Hex(expectedZip) != sha256Hex(vikunjaZip) {
		return nil, fmt.Errorf("vikunja ZIP is not the deterministic artifact for this input/config: expected_sha256=%s actual_sha256=%s", sha256Hex(expectedZip), sha256Hex(vikunjaZip))
	}
	reconciliation.Verified = true
	return reconciliation, nil
}
