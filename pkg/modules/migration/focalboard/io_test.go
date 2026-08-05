// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package focalboard

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWritePrivateFileDoesNotChmodExistingParent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o755))
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })
	require.NoError(t, WritePrivateFile("report.json", []byte("{}\n")))
	info, err := os.Stat(dir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestWritePrivateFileRefusesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	// #nosec G306 -- regression verifies caller-owned permissions are not changed.
	require.NoError(t, os.WriteFile(path, []byte("unrelated"), 0o644))
	err := WritePrivateFile(path, []byte("replacement"))
	require.ErrorContains(t, err, "refusing to overwrite")
	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	require.Equal(t, []byte("unrelated"), data)
	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	require.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestReadRegularFileLimitedRejectsOversizeAndSymlink(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input")
	require.NoError(t, os.WriteFile(file, []byte("12345"), 0o600))
	_, err := ReadRegularFileLimited(file, 4, "test input")
	require.ErrorContains(t, err, "exceeds")
	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink(file, link))
	_, err = ReadRegularFileLimited(link, 10, "test input")
	require.ErrorContains(t, err, "non-symlink")
}

func privateOutputTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o700))
	return dir
}

func TestWriteResultDoesNotChmodExistingOutputDirectory(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o755))
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "must already be private")
	info, statErr := os.Stat(dir)
	require.NoError(t, statErr)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestWriteResultPreflightLeavesExistingSetUntouched(t *testing.T) {
	dir := privateOutputTestDir(t)
	normalized := filepath.Join(dir, "normalized-focalboard.json")
	require.NoError(t, os.WriteFile(normalized, []byte("old"), 0o600))
	target := filepath.Join(dir, "outside")
	require.NoError(t, os.WriteFile(target, []byte("outside"), 0o600))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, "README.md")))
	result := ioTestResult(t, "new")
	err := WriteResult(dir, result)
	require.ErrorContains(t, err, "refusing to overwrite")
	data, readErr := os.ReadFile(normalized)
	require.NoError(t, readErr)
	require.Equal(t, []byte("old"), data)
	outside, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	require.Equal(t, []byte("outside"), outside)
}

func ioTestResult(t *testing.T, prefix string) *Result {
	t.Helper()
	result := &Result{NormalizedJSON: []byte(prefix + "-normalized"), VikunjaZip: []byte(prefix + "-zip"), ReconciliationJSON: []byte(prefix + "-json"), ReconciliationCSV: []byte(prefix + "-csv"), AssigneeTemplate: []byte(prefix + "-yaml"), README: []byte(prefix + "-readme")}
	manifest := RunManifest{SchemaVersion: NormalizedSchemaVersion, ToolVersion: ConverterVersion, ArtifactSHA256: map[string]string{"normalized-focalboard.json": sha256Hex(result.NormalizedJSON), "vikunja-import.zip": sha256Hex(result.VikunjaZip), "reconciliation.json": sha256Hex(result.ReconciliationJSON), "reconciliation.csv": sha256Hex(result.ReconciliationCSV), "assignees-map.template.yaml": sha256Hex(result.AssigneeTemplate), "README.md": sha256Hex(result.README)}}
	var err error
	result.ManifestJSON, err = marshalPretty(&manifest)
	require.NoError(t, err)
	return result
}

func TestWriteResultRollsBackPublishedSetOnRenameFailure(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	oldByName := map[string][]byte{"normalized-focalboard.json": oldResult.NormalizedJSON, "vikunja-import.zip": oldResult.VikunjaZip, "reconciliation.json": oldResult.ReconciliationJSON, "reconciliation.csv": oldResult.ReconciliationCSV, "assignees-map.template.yaml": oldResult.AssigneeTemplate, "README.md": oldResult.README, "RUN_MANIFEST.json": oldResult.ManifestJSON}
	originalRename := outputRename
	outputRename = func(oldPath, newPath string) error {
		if strings.Contains(oldPath, ".focalboard-migrate-staging-") && filepath.Base(newPath) == "reconciliation.json" {
			return errors.New("injected publish failure")
		}
		return os.Rename(oldPath, newPath)
	}
	t.Cleanup(func() { outputRename = originalRename })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "injected publish failure")
	for name, want := range oldByName {
		data, readErr := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, readErr)
		require.Equal(t, want, data)
	}
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	require.Len(t, entries, len(oldByName))
}

func TestWriteResultRetainsBackupWhenRollbackRestoreFails(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	originalRename := outputRename
	originalRestore := outputRestore
	outputRename = func(oldPath, newPath string) error {
		if strings.Contains(oldPath, ".focalboard-migrate-staging-") && filepath.Base(newPath) == "reconciliation.json" {
			return errors.New("injected publish failure")
		}
		return os.Rename(oldPath, newPath)
	}
	outputRestore = func(backupPath, targetPath string) error {
		if filepath.Base(targetPath) == "vikunja-import.zip" {
			return errors.New("injected restore failure")
		}
		return originalRestore(backupPath, targetPath)
	}
	t.Cleanup(func() { outputRename = originalRename; outputRestore = originalRestore })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorIs(t, err, errOutputRollbackIncomplete)
	require.ErrorContains(t, err, "injected restore failure")
	require.ErrorContains(t, err, "recovery backup retained")
	require.NoFileExists(t, filepath.Join(dir, "vikunja-import.zip"))
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	var backup string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".focalboard-migrate-backup-") {
			backup = filepath.Join(dir, entry.Name())
		}
	}
	require.NotEmpty(t, backup)
	data, readErr := os.ReadFile(filepath.Join(backup, "vikunja-import.zip"))
	require.NoError(t, readErr)
	require.Equal(t, oldResult.VikunjaZip, data)
}

func TestWriteResultRefusesUnrelatedExistingFiles(t *testing.T) {
	dir := privateOutputTestDir(t)
	readme := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(readme, []byte("repository readme"), 0o600))
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "refusing to overwrite")
	data, readErr := os.ReadFile(readme)
	require.NoError(t, readErr)
	require.Equal(t, []byte("repository readme"), data)
}

type injectedSyncDirectory struct {
	syncableDirectory
	syncErr error
}

func (d *injectedSyncDirectory) Sync() error { return d.syncErr }

func assertResultFiles(t *testing.T, dir string, result *Result) {
	t.Helper()
	want := map[string][]byte{
		"normalized-focalboard.json":  result.NormalizedJSON,
		"vikunja-import.zip":          result.VikunjaZip,
		"reconciliation.json":         result.ReconciliationJSON,
		"reconciliation.csv":          result.ReconciliationCSV,
		"assignees-map.template.yaml": result.AssigneeTemplate,
		"README.md":                   result.README,
		"RUN_MANIFEST.json":           result.ManifestJSON,
	}
	for name, expected := range want {
		actual, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		require.Equal(t, expected, actual, name)
	}
}

func findBackupDirectory(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".focalboard-migrate-backup-") {
			return filepath.Join(dir, entry.Name())
		}
	}
	return ""
}

func TestWriteResultFailsClosedOnBackupLstatError(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	originalLstat := outputLstat
	calls := 0
	outputLstat = func(name string) (os.FileInfo, error) {
		if filepath.Base(name) == "vikunja-import.zip" {
			calls++
			if calls == 3 {
				return nil, errors.New("injected backup lstat failure")
			}
		}
		return originalLstat(name)
	}
	t.Cleanup(func() { outputLstat = originalLstat })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "injected backup lstat failure")
	assertResultFiles(t, dir, oldResult)
	require.Empty(t, findBackupDirectory(t, dir))
}

func TestWriteResultDurablyRestoresAfterPublishSyncFailure(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	originalOpen := outputOpenDirectory
	dirSyncCalls := 0
	outputOpenDirectory = func(name string) (syncableDirectory, error) {
		directory, err := originalOpen(name)
		if err != nil {
			return nil, err
		}
		if name == dir {
			dirSyncCalls++
			if dirSyncCalls == 2 {
				return &injectedSyncDirectory{syncableDirectory: directory, syncErr: errors.New("injected publish sync failure")}, nil
			}
		}
		return directory, nil
	}
	t.Cleanup(func() { outputOpenDirectory = originalOpen })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "injected publish sync failure")
	require.NotErrorIs(t, err, errOutputRollbackIncomplete)
	assertResultFiles(t, dir, oldResult)
	require.Empty(t, findBackupDirectory(t, dir))
	require.GreaterOrEqual(t, dirSyncCalls, 4)
}

func TestWriteResultRetainsCompleteBackupWhenRollbackSyncFails(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	originalOpen := outputOpenDirectory
	dirSyncCalls := 0
	outputOpenDirectory = func(name string) (syncableDirectory, error) {
		directory, err := originalOpen(name)
		if err != nil {
			return nil, err
		}
		if name == dir {
			dirSyncCalls++
			if dirSyncCalls >= 2 {
				return &injectedSyncDirectory{syncableDirectory: directory, syncErr: errors.New("injected persistent sync failure")}, nil
			}
		}
		return directory, nil
	}
	t.Cleanup(func() { outputOpenDirectory = originalOpen })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorIs(t, err, errOutputRollbackIncomplete)
	require.ErrorContains(t, err, "recovery backup retained")
	backup := findBackupDirectory(t, dir)
	require.NotEmpty(t, backup)
	assertResultFiles(t, backup, oldResult)
}

func TestWriteResultDurablyRestoresAfterDirectoryOpenFailure(t *testing.T) {
	dir := privateOutputTestDir(t)
	oldResult := ioTestResult(t, "old")
	require.NoError(t, WriteResult(dir, oldResult))
	originalOpen := outputOpenDirectory
	dirOpenCalls := 0
	outputOpenDirectory = func(name string) (syncableDirectory, error) {
		if name == dir {
			dirOpenCalls++
			if dirOpenCalls == 2 {
				return nil, errors.New("injected directory open failure")
			}
		}
		return originalOpen(name)
	}
	t.Cleanup(func() { outputOpenDirectory = originalOpen })
	err := WriteResult(dir, ioTestResult(t, "new"))
	require.ErrorContains(t, err, "injected directory open failure")
	assertResultFiles(t, dir, oldResult)
	require.Empty(t, findBackupDirectory(t, dir))
}

func TestWriteResultRejectsInconsistentManifestBeforeWrites(t *testing.T) {
	dir := privateOutputTestDir(t)
	result := ioTestResult(t, "new")
	result.NormalizedJSON = []byte("tampered")
	err := WriteResult(dir, result)
	require.ErrorContains(t, err, "manifest hash mismatch")
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	require.Empty(t, entries)
}

func TestWriteResultRejectsNilResult(t *testing.T) {
	dir := privateOutputTestDir(t)
	require.ErrorContains(t, WriteResult(dir, nil), "must not be nil")
}
