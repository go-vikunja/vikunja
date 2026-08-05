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
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/utils"
)

const (
	maxInputSize          = 64 << 20
	maxNestedEntrySize    = 64 << 20
	maxNestedTotalSize    = 256 << 20
	maxOuterArchiveFiles  = 1024
	maxNestedArchiveFiles = 4096
)

var errStrictModeRequired = errors.New("strict mode is required for Focalboard migration")

type archiveRecord struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type sourceVersionRecord struct {
	Version int   `json:"version"`
	Date    int64 `json:"date"`
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func readLimited(r io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(r, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("archive entry exceeds %d bytes", limit)
	}
	return data, nil
}

func validateZipEntryName(name string) error {
	if name == "" || strings.ContainsRune(name, '\x00') || strings.HasPrefix(name, "/") || path.IsAbs(name) || utils.ContainsPathTraversal(name) {
		return fmt.Errorf("unsafe path in zip archive: %q", name)
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("unsafe path in zip archive: %q", name)
	}
	return nil
}

func parsePackage(inputPath string, opts Options) (*parsedArchive, error) {
	if !opts.Strict {
		return nil, errStrictModeRequired
	}
	data, err := ReadRegularFileLimited(inputPath, maxInputSize, "input archive")
	if err != nil {
		return nil, err
	}
	sourceHash := sha256Hex(data)
	if opts.ExpectedSHA256 != "" && sourceHash != strings.ToLower(opts.ExpectedSHA256) {
		return nil, fmt.Errorf("source SHA-256 mismatch: expected %s, got %s", strings.ToLower(opts.ExpectedSHA256), sourceHash)
	}

	outer, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open package ZIP: %w", err)
	}
	if len(outer.File) > maxOuterArchiveFiles {
		return nil, fmt.Errorf("package ZIP has too many entries: %d", len(outer.File))
	}
	var nestedFile *zip.File
	for _, f := range outer.File {
		if err := validateZipEntryName(f.Name); err != nil {
			return nil, err
		}
		if strings.HasPrefix(f.Name, "__MACOSX/") || strings.HasSuffix(f.Name, "/") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(f.Name), ".boardarchive") {
			if nestedFile != nil {
				return nil, fmt.Errorf("package contains more than one .boardarchive")
			}
			nestedFile = f
		}
	}
	if nestedFile == nil {
		return nil, fmt.Errorf("package does not contain a .boardarchive")
	}
	if nestedFile.UncompressedSize64 > maxInputSize {
		return nil, fmt.Errorf("nested .boardarchive exceeds %d bytes", maxInputSize)
	}
	r, err := nestedFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open nested .boardarchive: %w", err)
	}
	nestedBytes, readErr := readLimited(r, maxInputSize)
	closeErr := r.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read nested .boardarchive: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close nested .boardarchive: %w", closeErr)
	}

	nested, err := zip.NewReader(bytes.NewReader(nestedBytes), int64(len(nestedBytes)))
	if err != nil {
		return nil, fmt.Errorf("open nested .boardarchive ZIP: %w", err)
	}
	if len(nested.File) > maxNestedArchiveFiles {
		return nil, fmt.Errorf("nested .boardarchive has too many entries: %d", len(nested.File))
	}
	parsed := &parsedArchive{SourceHash: sourceHash, NestedHash: sha256Hex(nestedBytes)}
	var total uint64
	boardFiles := 0
	versionFiles := 0
	boardIDs := map[string]struct{}{}
	blockIDs := map[string]struct{}{}
	for _, f := range nested.File {
		if err := validateZipEntryName(f.Name); err != nil {
			return nil, err
		}
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		if f.UncompressedSize64 > maxNestedEntrySize {
			return nil, fmt.Errorf("nested entry %q exceeds size limit", f.Name)
		}
		total += f.UncompressedSize64
		if total > maxNestedTotalSize {
			return nil, fmt.Errorf("nested archive exceeds total uncompressed size limit")
		}
		if f.Name == "version.json" {
			versionFiles++
			if versionFiles > 1 {
				return nil, fmt.Errorf("nested archive contains duplicate version.json")
			}
			versionData, err := readZipEntryLimited(f, 1<<20)
			if err != nil {
				return nil, fmt.Errorf("read version.json: %w", err)
			}
			decoder := json.NewDecoder(bytes.NewReader(versionData))
			decoder.DisallowUnknownFields()
			var version sourceVersionRecord
			if err := decoder.Decode(&version); err != nil {
				return nil, fmt.Errorf("decode version.json: %w", err)
			}
			if decoder.Decode(&struct{}{}) != io.EOF {
				return nil, fmt.Errorf("decode version.json: trailing JSON data")
			}
			if version.Version != SupportedFocalboardVersion {
				return nil, fmt.Errorf("unsupported Focalboard archive version %d; expected %d", version.Version, SupportedFocalboardVersion)
			}
			if version.Date <= 0 {
				return nil, fmt.Errorf("version.json has invalid export date")
			}
			parsed.SourceVersion = version.Version
			continue
		}
		if !strings.HasSuffix(f.Name, "/board.jsonl") {
			parsed.Attachments++
			continue
		}
		boardFiles++
		if err := parseBoardJSONL(f, parsed, boardIDs, blockIDs, opts.Strict); err != nil {
			return nil, err
		}
	}
	if versionFiles != 1 || parsed.SourceVersion == 0 {
		return nil, fmt.Errorf("nested archive must contain exactly one valid version.json")
	}
	if boardFiles == 0 || len(parsed.Boards) == 0 {
		return nil, fmt.Errorf("nested archive contains no board.jsonl data")
	}
	if len(parsed.Boards) != boardFiles {
		return nil, fmt.Errorf("expected one board record per board.jsonl: files=%d boards=%d", boardFiles, len(parsed.Boards))
	}
	if opts.Strict && parsed.Attachments > 0 {
		return nil, fmt.Errorf("attachments are not supported by this converter: found %d", parsed.Attachments)
	}
	seenMembers := map[string]struct{}{}
	for _, member := range parsed.BoardMembers {
		if _, exists := boardIDs[member.BoardID]; !exists {
			return nil, fmt.Errorf("board member references unknown board %s", member.BoardID)
		}
		key := member.BoardID + "\x00" + member.UserID
		if _, exists := seenMembers[key]; exists {
			return nil, fmt.Errorf("duplicate board member for board %s", member.BoardID)
		}
		seenMembers[key] = struct{}{}
	}
	for _, group := range [][]*rawBlock{parsed.Cards, parsed.Texts, parsed.Views} {
		for _, block := range group {
			if _, exists := boardIDs[block.BoardID]; !exists {
				return nil, fmt.Errorf("block %s references unknown board %s", block.ID, block.BoardID)
			}
			if block.ParentID == "" {
				return nil, fmt.Errorf("block %s has empty parentId", block.ID)
			}
			if (block.Type == "card" || block.Type == "view") && block.ParentID != block.BoardID {
				return nil, fmt.Errorf("%s block %s parentId must equal boardId", block.Type, block.ID)
			}
			if block.ParentID == block.ID {
				return nil, fmt.Errorf("block %s cannot parent itself", block.ID)
			}
		}
	}

	sort.Slice(parsed.Boards, func(i, j int) bool { return parsed.Boards[i].ID < parsed.Boards[j].ID })
	sort.Slice(parsed.Cards, func(i, j int) bool {
		if parsed.Cards[i].BoardID == parsed.Cards[j].BoardID {
			if parsed.Cards[i].CreateAt == parsed.Cards[j].CreateAt {
				return parsed.Cards[i].ID < parsed.Cards[j].ID
			}
			return parsed.Cards[i].CreateAt < parsed.Cards[j].CreateAt
		}
		return parsed.Cards[i].BoardID < parsed.Cards[j].BoardID
	})
	sort.Slice(parsed.Texts, func(i, j int) bool {
		if parsed.Texts[i].BoardID == parsed.Texts[j].BoardID {
			if parsed.Texts[i].CreateAt == parsed.Texts[j].CreateAt {
				return parsed.Texts[i].ID < parsed.Texts[j].ID
			}
			return parsed.Texts[i].CreateAt < parsed.Texts[j].CreateAt
		}
		return parsed.Texts[i].BoardID < parsed.Texts[j].BoardID
	})
	return parsed, nil
}

func readZipEntryLimited(f *zip.File, limit int64) ([]byte, error) {
	if f.UncompressedSize64 > uint64(limit) {
		return nil, fmt.Errorf("zip entry %q exceeds %d bytes", f.Name, limit)
	}
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	data, readErr := readLimited(r, limit)
	closeErr := r.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, isDelim := token.(json.Delim)
		if !isDelim {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("JSON object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("source JSON contains duplicate key %q", key)
				}
				seen[key] = struct{}{}
				if err := walk(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walk(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delim)
		}
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("source JSON contains trailing data")
		}
		return err
	}
	return nil
}

func unmarshalSourceJSON(data []byte, target any, strict bool) error {
	if strict {
		if err := rejectDuplicateJSONKeys(data); err != nil {
			return err
		}
	}
	if !strict {
		return json.Unmarshal(data, target)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("source JSON contains trailing data")
		}
		return err
	}
	return nil
}

func parseBoardJSONL(f *zip.File, parsed *parsedArchive, boardIDs, blockIDs map[string]struct{}, strict bool) error {
	r, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %q: %w", f.Name, err)
	}
	defer r.Close()

	scanner := bufio.NewScanner(io.LimitReader(r, maxNestedEntrySize+1))
	scanner.Buffer(make([]byte, 64*1024), int(maxNestedEntrySize))
	line := 0
	for scanner.Scan() {
		line++
		var record archiveRecord
		if err := unmarshalSourceJSON(scanner.Bytes(), &record, strict); err != nil {
			return fmt.Errorf("decode %s line %d: %w", f.Name, line, err)
		}
		switch record.Type {
		case "board":
			var board rawBoard
			if err := unmarshalSourceJSON(record.Data, &board, strict); err != nil {
				return fmt.Errorf("decode board in %s line %d: %w", f.Name, line, err)
			}
			if board.ID == "" {
				return fmt.Errorf("board in %s line %d has empty id", f.Name, line)
			}
			if board.Type != "P" {
				return fmt.Errorf("board %s has unsupported type %q", board.ID, board.Type)
			}
			if _, exists := boardIDs[board.ID]; exists {
				return fmt.Errorf("duplicate board id %s", board.ID)
			}
			boardIDs[board.ID] = struct{}{}
			parsed.Boards = append(parsed.Boards, &board)
			parsed.MaxUpdated = maxMillisTime(parsed.MaxUpdated, board.UpdateAt)
		case "block":
			var block rawBlock
			if err := unmarshalSourceJSON(record.Data, &block, strict); err != nil {
				return fmt.Errorf("decode block in %s line %d: %w", f.Name, line, err)
			}
			if block.ID == "" || block.BoardID == "" {
				return fmt.Errorf("block in %s line %d has empty id or boardId", f.Name, line)
			}
			if _, exists := blockIDs[block.ID]; exists {
				return fmt.Errorf("duplicate block id %s", block.ID)
			}
			blockIDs[block.ID] = struct{}{}
			parsed.MaxUpdated = maxMillisTime(parsed.MaxUpdated, block.UpdateAt)
			switch block.Type {
			case "card":
				parsed.Cards = append(parsed.Cards, &block)
			case "text":
				parsed.Texts = append(parsed.Texts, &block)
			case "view":
				if block.Fields.ViewType != "board" {
					return fmt.Errorf("unsupported Focalboard view type %q for view %s", block.Fields.ViewType, block.ID)
				}
				parsed.Views = append(parsed.Views, &block)
			default:
				return fmt.Errorf("unsupported Focalboard block type %q", block.Type)
			}
		case "boardMember":
			var member rawBoardMember
			if err := unmarshalSourceJSON(record.Data, &member, strict); err != nil {
				return fmt.Errorf("decode board member in %s line %d: %w", f.Name, line, err)
			}
			if member.BoardID == "" || member.UserID == "" {
				return fmt.Errorf("board member in %s line %d has empty boardId or userId", f.Name, line)
			}
			parsed.BoardMembers = append(parsed.BoardMembers, member)
			// Membership cannot be applied before Vikunja users exist.
		default:
			return fmt.Errorf("unsupported Focalboard record type %q", record.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", f.Name, err)
	}
	return nil
}

func maxMillisTime(current time.Time, millis int64) time.Time {
	if millis <= 0 {
		return current
	}
	candidate := time.UnixMilli(millis).UTC()
	if candidate.After(current) {
		return candidate
	}
	return current
}
