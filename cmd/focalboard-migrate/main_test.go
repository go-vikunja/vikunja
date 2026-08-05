// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidRepoSHA(t *testing.T) {
	require.True(t, validRepoSHA(strings.Repeat("a", 40)))
	require.True(t, validRepoSHA(strings.Repeat("b", 64)))
	require.False(t, validRepoSHA(""))
	require.False(t, validRepoSHA(strings.Repeat("z", 40)))
}

func TestRunRequiresCommandAndFlags(t *testing.T) {
	require.Error(t, run(nil))
	require.Error(t, run([]string{"unknown"}))
	require.Error(t, run([]string{"convert"}))
	require.Error(t, run([]string{"analyze"}))
	require.Error(t, run([]string{"verify"}))
}

func TestRejectOutputAliases(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.zip")
	require.NoError(t, os.WriteFile(input, []byte("input"), 0o600))
	require.Error(t, rejectOutputAliases(input, []string{input}))
	link := filepath.Join(dir, "alias")
	require.NoError(t, os.Symlink(input, link))
	require.Error(t, rejectOutputAliases(link, []string{input}))
	require.NoError(t, rejectOutputAliases(filepath.Join(dir, "report.json"), []string{input}))
}

func TestRejectOutputDirectoryContainingInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.zip")
	require.NoError(t, os.WriteFile(input, []byte("input"), 0o600))
	require.Error(t, rejectOutputDirectoryOverlap(dir, []string{input}))
	require.NoError(t, rejectOutputDirectoryOverlap(filepath.Join(t.TempDir(), "new-output"), []string{input}))
}

func TestCommandsRejectInputOutputAliasingBeforeRead(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "source.zip")
	require.NoError(t, os.WriteFile(input, []byte("sentinel-source"), 0o600))
	err := runAnalyze([]string{"--input", input, "--report", input})
	require.ErrorContains(t, err, "must not alias")
	data, readErr := os.ReadFile(input)
	require.NoError(t, readErr)
	require.Equal(t, []byte("sentinel-source"), data)
	err = runConvert([]string{"--input", input, "--output-dir", dir})
	require.ErrorContains(t, err, "must not contain input")
	data, readErr = os.ReadFile(input)
	require.NoError(t, readErr)
	require.Equal(t, []byte("sentinel-source"), data)
}

func TestRepositoryProvenanceIsBoundToActualVikunjaHead(t *testing.T) {
	root, err := findVikunjaRepoRoot()
	require.NoError(t, err)
	sha, _ := currentRepoState(root)
	require.True(t, validRepoSHA(sha))
	dir := t.TempDir()
	input := filepath.Join(dir, "source.zip")
	require.NoError(t, os.WriteFile(input, []byte("not-used"), 0o600))
	err = runConvert([]string{"--input", input, "--output-dir", filepath.Join(dir, "out"), "--repo-sha", strings.Repeat("0", 40)})
	require.ErrorContains(t, err, "does not match Vikunja repository HEAD")
}

func TestFindVikunjaRepoRootRejectsNestedModuleInParentGitRepo(t *testing.T) {
	parent := t.TempDir()
	require.NoError(t, exec.CommandContext(context.Background(), "git", "init", parent).Run())
	nested := filepath.Join(parent, "nested")
	require.NoError(t, os.MkdirAll(nested, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "go.mod"), []byte("module code.vikunja.io/api\n"), 0o600))
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(nested))
	t.Cleanup(func() { _ = os.Chdir(old) })
	_, err = findVikunjaRepoRoot()
	require.ErrorIs(t, err, errNotVikunjaRepo)
	sha, dirty := currentRepoState(nested)
	require.Empty(t, sha)
	require.True(t, dirty)
}

func TestExactModuleDeclarationDoesNotMatchCommentsOrPrefixes(t *testing.T) {
	require.False(t, hasExactVikunjaModule([]byte("// module code.vikunja.io/api\nmodule example.com/fake\n")))
	require.False(t, hasExactVikunjaModule([]byte("module code.vikunja.io/api-fake\n")))
	require.True(t, hasExactVikunjaModule([]byte("module code.vikunja.io/api\n")))
}

func TestCommandsRequireStrictModeAndSourceHash(t *testing.T) {
	require.ErrorIs(t, runAnalyze([]string{"--strict=false"}), errStrictRequired)
	require.ErrorIs(t, runConvert([]string{"--strict=false"}), errStrictRequired)
	require.ErrorIs(t, runVerify([]string{"--strict=false"}), errStrictRequired)
	require.ErrorContains(t, runAnalyze([]string{"--expected-sha256="}), "--expected-sha256")
}

func TestHasExactVikunjaModuleUsesGoModSyntax(t *testing.T) {
	require.True(t, hasExactVikunjaModule([]byte("module code.vikunja.io/api\n\ngo 1.26.4\n")))
	require.False(t, hasExactVikunjaModule([]byte("/*\nmodule code.vikunja.io/api\n*/\nmodule example.invalid/other\n\ngo 1.26.4\n")))
	require.False(t, hasExactVikunjaModule([]byte("module example.invalid/other // module code.vikunja.io/api\n\ngo 1.26.4\n")))
}
