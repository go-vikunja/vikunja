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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.vikunja.io/api/pkg/modules/migration/focalboard"

	"golang.org/x/mod/modfile"
)

var (
	errUsage               = errors.New("usage: focalboard-migrate <analyze|convert|verify> [flags]")
	errUnknownCommand      = errors.New("unknown command")
	errRequiredFlag        = errors.New("required flag is missing")
	errInvalidRepoSHA      = errors.New("repo SHA must be a 40- or 64-character hexadecimal git object id; auto-detection also failed")
	errStrictRequired      = errors.New("strict mode is mandatory for migration artifacts")
	errRepoSHAMismatch     = errors.New("--repo-sha does not match Vikunja repository HEAD")
	errCannotResolvePath   = errors.New("cannot resolve path")
	errOutputAliasesInput  = errors.New("output path must not alias input path")
	errOutputContainsInput = errors.New("output directory must not contain input path")
	errNotVikunjaRepo      = errors.New("must run from the Vikunja repository rooted at module code.vikunja.io/api")
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "focalboard-migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "analyze":
		return runAnalyze(args[1:])
	case "convert":
		return runConvert(args[1:])
	case "verify":
		return runVerify(args[1:])
	default:
		return fmt.Errorf("%w: %q", errUnknownCommand, args[0])
	}
}

func commonFlags(fs *flag.FlagSet) (input, expectedSHA, timezone, version *string, strict *bool) {
	input = fs.String("input", "", "Focalboard package ZIP")
	expectedSHA = fs.String("expected-sha256", focalboard.KnownSourceSHA256, "required source archive SHA-256")
	timezone = fs.String("timezone", "", "IANA timezone for strict ISO due dates")
	version = fs.String("vikunja-version", focalboard.DefaultVikunjaVersion, "VERSION entry for native import ZIP")
	strict = fs.Bool("strict", true, "fail closed on ambiguous or unsupported input")
	return
}

func options(expectedSHA, timezone, version string, strict bool, assignees *focalboard.AssigneeMap, repoSHA string, repoDirty bool) focalboard.Options {
	return focalboard.Options{
		ExpectedSHA256: expectedSHA,
		Strict:         strict,
		Timezone:       timezone,
		VikunjaVersion: version,
		Assignees:      assignees,
		RepoSHA:        repoSHA,
		RepoDirty:      repoDirty,
	}
}

func requireFlag(name, value string) error {
	if value == "" {
		return fmt.Errorf("%w: --%s", errRequiredFlag, name)
	}
	return nil
}

func runAnalyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	input, expectedSHA, timezone, version, strict := commonFlags(fs)
	report := fs.String("report", "", "analysis report JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*strict {
		return errStrictRequired
	}
	if err := requireFlag("expected-sha256", *expectedSHA); err != nil {
		return err
	}
	if err := requireFlag("input", *input); err != nil {
		return err
	}
	if err := requireFlag("report", *report); err != nil {
		return err
	}
	if err := rejectOutputAliases(*report, []string{*input}); err != nil {
		return err
	}
	analysis, err := focalboard.Analyze(*input, options(*expectedSHA, *timezone, *version, *strict, nil, "", false))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := focalboard.WritePrivateFile(*report, data); err != nil {
		return err
	}
	return printSummary(map[string]any{
		"source_archive_sha256": analysis.SourceArchiveSHA256,
		"counts":                analysis.Counts,
		"warning_count":         len(analysis.Warnings),
	})
}

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	input, expectedSHA, timezone, version, strict := commonFlags(fs)
	outputDir := fs.String("output-dir", "", "private output directory")
	assigneesFile := fs.String("assignees-map", "", "optional assignee mapping YAML or JSON")
	repoSHA := fs.String("repo-sha", "", "repository commit SHA for RUN_MANIFEST")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*strict {
		return errStrictRequired
	}
	if err := requireFlag("expected-sha256", *expectedSHA); err != nil {
		return err
	}
	if err := requireFlag("input", *input); err != nil {
		return err
	}
	if err := requireFlag("output-dir", *outputDir); err != nil {
		return err
	}
	inputPaths := []string{*input}
	if *assigneesFile != "" {
		inputPaths = append(inputPaths, *assigneesFile)
	}
	if err := rejectOutputDirectoryOverlap(*outputDir, inputPaths); err != nil {
		return err
	}
	mapping, err := focalboard.LoadAssigneeMap(*assigneesFile)
	if err != nil {
		return err
	}
	repoRoot, err := findVikunjaRepoRoot()
	if err != nil {
		return err
	}
	detectedSHA, repoDirty := currentRepoState(repoRoot)
	if !validRepoSHA(detectedSHA) {
		return errInvalidRepoSHA
	}
	if *repoSHA != "" && *repoSHA != detectedSHA {
		return fmt.Errorf("%w: supplied=%s detected=%s", errRepoSHAMismatch, *repoSHA, detectedSHA)
	}
	*repoSHA = detectedSHA
	result, err := focalboard.Convert(*input, options(*expectedSHA, *timezone, *version, *strict, mapping, *repoSHA, repoDirty))
	if err != nil {
		return err
	}
	if err := focalboard.WriteResult(*outputDir, result); err != nil {
		return err
	}
	return printSummary(map[string]any{
		"run_id":                  result.Manifest.RunID,
		"source_archive_sha256":   result.Manifest.SourceArchiveSHA256,
		"counts":                  result.Manifest.Counts,
		"artifact_sha256":         result.Manifest.ArtifactSHA256,
		"reconciliation_verified": result.Reconciliation.Verified,
	})
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	input, expectedSHA, timezone, version, strict := commonFlags(fs)
	normalizedPath := fs.String("normalized", "", "normalized JSON")
	vikunjaZipPath := fs.String("vikunja-zip", "", "native Vikunja import ZIP")
	report := fs.String("report", "", "reconciliation report JSON")
	assigneesFile := fs.String("assignees-map", "", "optional assignee mapping YAML or JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*strict {
		return errStrictRequired
	}
	if err := requireFlag("expected-sha256", *expectedSHA); err != nil {
		return err
	}
	for name, value := range map[string]string{"input": *input, "normalized": *normalizedPath, "vikunja-zip": *vikunjaZipPath, "report": *report} {
		if err := requireFlag(name, value); err != nil {
			return err
		}
	}
	inputPaths := []string{*input, *normalizedPath, *vikunjaZipPath}
	if *assigneesFile != "" {
		inputPaths = append(inputPaths, *assigneesFile)
	}
	if err := rejectOutputAliases(*report, inputPaths); err != nil {
		return err
	}
	mapping, err := focalboard.LoadAssigneeMap(*assigneesFile)
	if err != nil {
		return err
	}
	normalizedJSON, err := focalboard.ReadRegularFileLimited(*normalizedPath, focalboard.MaxNormalizedJSONSize, "normalized JSON")
	if err != nil {
		return err
	}
	vikunjaZip, err := focalboard.ReadRegularFileLimited(*vikunjaZipPath, focalboard.MaxVikunjaZipSize, "Vikunja ZIP")
	if err != nil {
		return err
	}
	reconciliation, err := focalboard.Verify(*input, normalizedJSON, vikunjaZip, options(*expectedSHA, *timezone, *version, *strict, mapping, "", false))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(reconciliation, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := focalboard.WritePrivateFile(*report, data); err != nil {
		return err
	}
	return printSummary(map[string]any{
		"verified":              reconciliation.Verified,
		"source_archive_sha256": reconciliation.SourceArchiveSHA256,
		"counts":                reconciliation.Counts,
		"normalized_sha256":     sha256ForSummary(normalizedJSON),
		"vikunja_zip_sha256":    sha256ForSummary(vikunjaZip),
	})
}

func canonicalPath(filename string) (string, error) {
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	current := filepath.Clean(absolute)
	missing := []string{}
	for {
		if _, err := os.Lstat(current); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("%w: %s", errCannotResolvePath, filename)
		}
		missing = append([]string{filepath.Base(current)}, missing...)
		current = parent
	}
	resolved, err := filepath.EvalSymlinks(current)
	if err != nil {
		return "", err
	}
	parts := append([]string{resolved}, missing...)
	return filepath.Clean(filepath.Join(parts...)), nil
}

func rejectOutputAliases(output string, inputs []string) error {
	canonicalOutput, err := canonicalPath(output)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		canonicalInput, err := canonicalPath(input)
		if err != nil {
			return err
		}
		if canonicalOutput == canonicalInput {
			return fmt.Errorf("%w: %s", errOutputAliasesInput, input)
		}
	}
	return nil
}

func rejectOutputDirectoryOverlap(outputDir string, inputs []string) error {
	canonicalDir, err := canonicalPath(outputDir)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		canonicalInput, err := canonicalPath(input)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(canonicalDir, canonicalInput)
		if err != nil {
			return err
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))) {
			return fmt.Errorf("%w: %s", errOutputContainsInput, input)
		}
	}
	return nil
}

func hasExactVikunjaModule(data []byte) bool {
	parsed, err := modfile.Parse("go.mod", data, nil)
	return err == nil && parsed.Module != nil && parsed.Module.Mod.Path == "code.vikunja.io/api"
}

func gitTopLevel(dir string) (string, error) {
	command := exec.CommandContext(context.Background(), "git", "-C", dir, "rev-parse", "--show-toplevel")
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return canonicalPath(strings.TrimSpace(string(output)))
}

func exactGitRoot(dir string) bool {
	candidate, err := canonicalPath(dir)
	if err != nil {
		return false
	}
	top, err := gitTopLevel(dir)
	return err == nil && top == candidate
}

func findVikunjaRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		goMod := filepath.Join(dir, "go.mod")
		data, readErr := focalboard.ReadRegularFileLimited(goMod, 1<<20, "Vikunja go.mod")
		if readErr == nil && hasExactVikunjaModule(data) {
			if !exactGitRoot(dir) {
				return "", errNotVikunjaRepo
			}
			return canonicalPath(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errNotVikunjaRepo
}

func validRepoSHA(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func currentRepoState(repoRoot string) (sha string, dirty bool) {
	if !exactGitRoot(repoRoot) {
		return "", true
	}
	command := exec.CommandContext(context.Background(), "git", "-C", repoRoot, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil {
		return "", true
	}
	status := exec.CommandContext(context.Background(), "git", "-C", repoRoot, "status", "--porcelain", "--untracked-files=all")
	statusOutput, statusErr := status.Output()
	return strings.TrimSpace(string(output)), statusErr != nil || len(statusOutput) > 0
}

func sha256ForSummary(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func printSummary(summary any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(summary)
}
