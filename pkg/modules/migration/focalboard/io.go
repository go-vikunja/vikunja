// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package focalboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	MaxAssigneeMapSize    int64 = 4 << 20
	MaxNormalizedJSONSize int64 = 64 << 20
	MaxVikunjaZipSize     int64 = 256 << 20
)

func ensurePrivateOutputDirectory(dir string) error {
	if dir == "" {
		return fmt.Errorf("output directory is required")
	}
	if info, err := os.Lstat(dir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("output path is not a real directory")
		}
		if info.Mode().Perm()&0o077 != 0 {
			return fmt.Errorf("existing output directory must already be private (mode 0700 or stricter)")
		}
		return nil // Never chmod a caller-owned existing directory.
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return os.Chmod(dir, 0o700)
}

func ensureParentDirectory(dir string) error {
	if info, err := os.Lstat(dir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("parent path is not a real directory")
		}
		return nil // Never chmod a caller-owned existing parent.
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	return os.Chmod(dir, 0o700)
}

// ReadRegularFileLimited validates and reads the same regular-file descriptor.
func ReadRegularFileLimited(filename string, limit int64, label string) ([]byte, error) {
	before, err := os.Lstat(filename)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", label, err)
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular non-symlink file", label)
	}
	file, err := os.Open(filename) // #nosec G304 -- explicit CLI input, descriptor validated below
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", label, err)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("fstat %s: %w", label, err)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		return nil, fmt.Errorf("%s changed before it was opened", label)
	}
	if opened.Size() > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes", label, limit)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes", label, limit)
	}
	after, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("fstat %s after read: %w", label, err)
	}
	if !os.SameFile(opened, after) || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
		return nil, fmt.Errorf("%s changed while it was being read", label)
	}
	return data, nil
}

func WritePrivateFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := ensureParentDirectory(dir); err != nil {
		return err
	}
	if info, err := os.Lstat(filename); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to replace non-regular file %s", filename)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".focalboard-migrate-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Link(tmpName, filename); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("refusing to overwrite existing file %s: %w", filename, err)
		}
		return err
	}
	if err := os.Remove(tmpName); err != nil {
		return fmt.Errorf("remove temporary hard link: %w", err)
	}
	parent, err := os.Open(filepath.Dir(filename)) // #nosec G304 -- explicit caller-selected parent
	if err != nil {
		return err
	}
	syncErr := parent.Sync()
	closeErr := parent.Close()
	return errors.Join(syncErr, closeErr)
}

type syncableDirectory interface {
	Sync() error
	Close() error
}

var (
	outputRename                = os.Rename
	outputRemove                = os.Remove
	outputRemoveAll             = os.RemoveAll
	outputLstat                 = os.Lstat
	outputOpenDirectory         = func(name string) (syncableDirectory, error) { return os.Open(name) }
	outputRestore               = restoreOutputFile
	errOutputRollbackIncomplete = errors.New("output publication rollback incomplete")
)

func syncOutputDirectory(dir string) error {
	directory, err := outputOpenDirectory(dir)
	if err != nil {
		return fmt.Errorf("open directory %s: %w", dir, err)
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	return errors.Join(syncErr, closeErr)
}

func restoreOutputFile(backupPath, targetPath string) error {
	data, err := ReadRegularFileLimited(backupPath, MaxVikunjaZipSize, "backup output artifact")
	if err != nil {
		return err
	}
	return WritePrivateFile(targetPath, data)
}

func rollbackPublishedSet(dir, backup string, published, backedUp []string) error {
	var rollbackErrors []error
	for i := len(published) - 1; i >= 0; i-- {
		name := published[i]
		if err := outputRemove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("remove published %s: %w", name, err))
		}
	}
	for i := len(backedUp) - 1; i >= 0; i-- {
		name := backedUp[i]
		if err := outputRestore(filepath.Join(backup, name), filepath.Join(dir, name)); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore %s: %w", name, err))
		}
	}
	return errors.Join(rollbackErrors...)
}

func publicationFailure(primary error, dir, backup string, published, backedUp []string) error {
	if rollbackErr := rollbackPublishedSet(dir, backup, published, backedUp); rollbackErr != nil {
		return fmt.Errorf("%w; operation error: %w; rollback error: %w; recovery backup retained at %s", errOutputRollbackIncomplete, primary, rollbackErr, backup)
	}
	if syncErr := syncOutputDirectory(dir); syncErr != nil {
		return fmt.Errorf("%w; operation error: %w; restored output sync failed: %w; recovery backup retained at %s", errOutputRollbackIncomplete, primary, syncErr, backup)
	}
	if err := outputRemoveAll(backup); err != nil {
		return fmt.Errorf("%w; operation error: %w; restored prior output but could not remove backup %s: %w", errOutputRollbackIncomplete, primary, backup, err)
	}
	if syncErr := syncOutputDirectory(dir); syncErr != nil {
		return fmt.Errorf("%w; operation error: %w; backup removal sync failed: %w", errOutputRollbackIncomplete, primary, syncErr)
	}
	return primary
}

func validateResultForWrite(result *Result) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}
	artifacts := map[string][]byte{
		"normalized-focalboard.json":  result.NormalizedJSON,
		"vikunja-import.zip":          result.VikunjaZip,
		"reconciliation.json":         result.ReconciliationJSON,
		"reconciliation.csv":          result.ReconciliationCSV,
		"assignees-map.template.yaml": result.AssigneeTemplate,
		"README.md":                   result.README,
	}
	for name, data := range artifacts {
		if len(data) == 0 {
			return fmt.Errorf("result artifact %s is empty", name)
		}
	}
	var manifest RunManifest
	if err := json.Unmarshal(result.ManifestJSON, &manifest); err != nil {
		return fmt.Errorf("decode result manifest: %w", err)
	}
	if manifest.SchemaVersion != NormalizedSchemaVersion || manifest.ToolVersion != ConverterVersion {
		return fmt.Errorf("result manifest has unsupported schema or tool version")
	}
	if len(manifest.ArtifactSHA256) != len(artifacts) {
		return fmt.Errorf("result manifest artifact set is incomplete")
	}
	for name, data := range artifacts {
		if manifest.ArtifactSHA256[name] != sha256Hex(data) {
			return fmt.Errorf("result manifest hash mismatch for %s", name)
		}
	}
	return nil
}

func WriteResult(dir string, result *Result) (returnErr error) {
	if err := validateResultForWrite(result); err != nil {
		return err
	}
	if err := ensurePrivateOutputDirectory(dir); err != nil {
		return err
	}
	files := []struct {
		name string
		data []byte
	}{
		{"normalized-focalboard.json", result.NormalizedJSON},
		{"vikunja-import.zip", result.VikunjaZip},
		{"reconciliation.json", result.ReconciliationJSON},
		{"reconciliation.csv", result.ReconciliationCSV},
		{"assignees-map.template.yaml", result.AssigneeTemplate},
		{"README.md", result.README},
		{"RUN_MANIFEST.json", result.ManifestJSON}, // completion marker, publish last
	}
	existingOutputs := 0
	for _, file := range files {
		if _, err := outputLstat(filepath.Join(dir, file.name)); err == nil {
			existingOutputs++
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("preflight %s: %w", file.name, err)
		}
	}
	if existingOutputs > 0 {
		if existingOutputs != len(files) {
			return fmt.Errorf("refusing to overwrite incomplete or unrelated output set")
		}
		manifestData, err := ReadRegularFileLimited(filepath.Join(dir, "RUN_MANIFEST.json"), 4<<20, "existing run manifest")
		if err != nil {
			return fmt.Errorf("refusing to overwrite unverified output set: %w", err)
		}
		var previous RunManifest
		if err := json.Unmarshal(manifestData, &previous); err != nil || previous.SchemaVersion != NormalizedSchemaVersion || previous.ToolVersion != ConverterVersion {
			return fmt.Errorf("refusing to overwrite output set without a valid run manifest")
		}
		for _, file := range files {
			if file.name == "RUN_MANIFEST.json" {
				continue
			}
			expected, exists := previous.ArtifactSHA256[file.name]
			if !exists {
				return fmt.Errorf("refusing to overwrite output set with incomplete artifact manifest")
			}
			oldData, err := ReadRegularFileLimited(filepath.Join(dir, file.name), MaxVikunjaZipSize, "existing output artifact")
			if err != nil || sha256Hex(oldData) != expected {
				return fmt.Errorf("refusing to overwrite output set with artifact hash mismatch")
			}
		}
	}

	for _, file := range files {
		target := filepath.Join(dir, file.name)
		if info, err := outputLstat(target); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return fmt.Errorf("refusing to replace non-regular output %s", file.name)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("preflight %s: %w", file.name, err)
		}
	}
	staging, err := os.MkdirTemp(dir, ".focalboard-migrate-staging-*")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(staging); cleanupErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove staging directory: %w", cleanupErr))
		}
	}()
	if err := os.Chmod(staging, 0o700); err != nil {
		return err
	}
	for _, file := range files {
		if err := WritePrivateFile(filepath.Join(staging, file.name), file.data); err != nil {
			return fmt.Errorf("stage %s: %w", file.name, err)
		}
	}
	if err := syncOutputDirectory(staging); err != nil {
		return fmt.Errorf("sync staging directory: %w", err)
	}
	backup, err := os.MkdirTemp(dir, ".focalboard-migrate-backup-*")
	if err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	if err := os.Chmod(backup, 0o700); err != nil {
		cleanupErr := outputRemoveAll(backup)
		if cleanupErr != nil {
			return fmt.Errorf("chmod backup directory: %w; cleanup failed: %w", err, cleanupErr)
		}
		return err
	}
	backedUp := make([]string, 0, len(files))
	for _, file := range files {
		target := filepath.Join(dir, file.name)
		if _, err := outputLstat(target); err == nil {
			if err := outputRename(target, filepath.Join(backup, file.name)); err != nil {
				return publicationFailure(fmt.Errorf("backup %s: %w", file.name, err), dir, backup, nil, backedUp)
			}
			backedUp = append(backedUp, file.name)
		} else if !os.IsNotExist(err) {
			return publicationFailure(fmt.Errorf("backup lstat %s: %w", file.name, err), dir, backup, nil, backedUp)
		}
	}
	if err := syncOutputDirectory(backup); err != nil {
		return publicationFailure(fmt.Errorf("sync backup directory: %w", err), dir, backup, nil, backedUp)
	}
	if err := syncOutputDirectory(dir); err != nil {
		return publicationFailure(fmt.Errorf("sync output directory after backup: %w", err), dir, backup, nil, backedUp)
	}
	published := make([]string, 0, len(files))
	for _, file := range files {
		if err := outputRename(filepath.Join(staging, file.name), filepath.Join(dir, file.name)); err != nil {
			return publicationFailure(fmt.Errorf("publish %s: %w", file.name, err), dir, backup, published, backedUp)
		}
		published = append(published, file.name)
	}
	if err := syncOutputDirectory(dir); err != nil {
		return publicationFailure(fmt.Errorf("sync published output directory: %w", err), dir, backup, published, backedUp)
	}
	if err := outputRemoveAll(backup); err != nil {
		return fmt.Errorf("remove completed backup directory: %w", err)
	}
	if err := syncOutputDirectory(dir); err != nil {
		return fmt.Errorf("sync completed output directory: %w", err)
	}
	return nil
}

func LoadAssigneeMap(filename string) (*AssigneeMap, error) {
	if filename == "" {
		return nil, nil
	}
	data, err := ReadRegularFileLimited(filename, MaxAssigneeMapSize, "assignee map")
	if err != nil {
		return nil, err
	}
	mapping := &AssigneeMap{}
	if strings.EqualFold(filepath.Ext(filename), ".json") {
		err = unmarshalSourceJSON(data, mapping, true)
	} else {
		err = yaml.UnmarshalWithOptions(data, mapping, yaml.DisallowUnknownField())
	}
	if err != nil {
		return nil, fmt.Errorf("decode assignee map: %w", err)
	}
	if mapping.SchemaVersion == "" {
		mapping.SchemaVersion = AssigneeMapSchemaVersion
	}
	if _, err := buildAssigneeLookup(mapping); err != nil {
		return nil, err
	}
	return mapping, nil
}
