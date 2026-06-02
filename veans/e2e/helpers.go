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

// Package e2e is the integration suite for veans. It assumes a running
// Vikunja API at VEANS_E2E_API_URL with VIKUNJA_SERVICE_TESTINGTOKEN set
// (passed in via VEANS_E2E_TESTING_TOKEN) so the suite can seed its own
// admin via PATCH /api/v1/test/users — the same `/test/{table}` endpoint
// the frontend playwright suite uses.
//
// The alternative path — VEANS_E2E_ADMIN_TOKEN — is a JWT against a
// long-lived Vikunja the user wants to drive without touching its data;
// in that mode the suite skips the seed.
//
// The suite never provisions Vikunja itself.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/veans/internal/client"
)

// Hard-coded seed credentials. The hash is the bcrypt of "1234" and
// matches frontend/tests/support/constants.ts so the whole e2e infra
// shares one well-known password — tests themselves never need to read
// these from env.
const (
	seedAdminUsername = "e2eadmin"
	seedAdminPassword = "1234"
	seedAdminBcrypt   = "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To." //nolint:gosec // G101: deterministic test fixture, not a credential
)

// Harness bundles a built veans binary and an authenticated admin client
// for verifying side effects on the server.
type Harness struct {
	Binary      string
	APIURL      string
	AdminToken  string
	AdminClient *client.Client
}

// New builds (or reuses) the veans binary, seeds the admin user via
// PATCH /api/v1/test/users (using VEANS_E2E_TESTING_TOKEN), logs in as
// that admin, and returns a Harness ready to drive tests.
//
// If VEANS_E2E_ADMIN_TOKEN is set, the seed is skipped and that token
// is used directly — useful for running against a long-lived Vikunja
// the caller doesn't want this suite to mutate user rows on.
//
// Tests rely on the `-short` skip in TestMain to opt out when a Vikunja
// instance isn't available; if `-short` is *not* set and env is missing,
// we fail loudly with a "configure or pass -short" hint.
func New(t *testing.T) *Harness {
	t.Helper()

	apiURL := strings.TrimRight(os.Getenv("VEANS_E2E_API_URL"), "/")
	if apiURL == "" {
		t.Fatal("VEANS_E2E_API_URL is not set — point it at a Vikunja instance, or pass -short to skip the e2e suite")
	}
	binary, err := buildOrLocate()
	if err != nil {
		t.Fatalf("locate veans binary: %v", err)
	}

	tok := os.Getenv("VEANS_E2E_ADMIN_TOKEN")
	if tok == "" {
		testingToken := os.Getenv("VEANS_E2E_TESTING_TOKEN")
		if testingToken == "" {
			t.Fatal("set VEANS_E2E_ADMIN_TOKEN, or VEANS_E2E_TESTING_TOKEN (matching the API's VIKUNJA_SERVICE_TESTINGTOKEN) so the suite can seed its own admin")
		}
		seedAdmin(t, apiURL, testingToken)
		c := client.New(apiURL, "")
		resp, err := c.Login(t.Context(), &client.LoginRequest{
			Username:  seedAdminUsername,
			Password:  seedAdminPassword,
			LongToken: true,
		})
		if err != nil {
			t.Fatalf("admin login: %v", err)
		}
		tok = resp.Token
	}

	return &Harness{
		Binary:      binary,
		APIURL:      apiURL,
		AdminToken:  tok,
		AdminClient: client.New(apiURL, tok),
	}
}

// seedAdmin PATCHes a single admin user row into the users table via
// the testing endpoint. truncate=true wipes any prior users from
// previous tests so each New(t) starts from a known state.
func seedAdmin(t *testing.T, apiURL, testingToken string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	body, err := json.Marshal([]map[string]any{{
		"id":       1,
		"username": seedAdminUsername,
		"password": seedAdminBcrypt,
		"email":    "e2e@example.com",
		"status":   0,
		"issuer":   "local",
		"language": "en",
		"created":  now,
		"updated":  now,
	}})
	if err != nil {
		t.Fatalf("marshal seed payload: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch,
		apiURL+"/api/v1/test/users?truncate=true", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build seed request: %v", err)
	}
	req.Header.Set("Authorization", testingToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		t.Fatalf("seed admin: HTTP %d: %s", resp.StatusCode, string(buf))
	}
}

// Workspace creates a per-test git repo in a TempDir with HOME pointed at
// a TempDir so the credential store writes under the test's own directory
// rather than touching the developer's keychain.
type Workspace struct {
	Dir          string
	Home         string
	ConfigPath   string
	BotUsername  string
	envOverrides map[string]string
}

// NewWorkspace initializes a fresh repo with `git init` + a single commit so
// `git rev-parse --abbrev-ref HEAD` returns a real branch name.
func (h *Harness) NewWorkspace(t *testing.T) *Workspace {
	t.Helper()
	dir := t.TempDir()
	home := t.TempDir()

	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "veans-e2e@example.com"},
		{"git", "config", "user.name", "veans-e2e"},
		// Disable any inherited commit signing; the test commit doesn't
		// need provenance and signing brokers can fail in dev containers.
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "config", "tag.gpgsign", "false"},
	} {
		cmd := exec.CommandContext(t.Context(), c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(c, " "), err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-q", "-m", "initial"},
	} {
		cmd := exec.CommandContext(t.Context(), c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(c, " "), err, out)
		}
	}

	return &Workspace{
		Dir:        dir,
		Home:       home,
		ConfigPath: filepath.Join(dir, ".veans.yml"),
		envOverrides: map[string]string{
			"HOME": home,
		},
	}
}

// Run executes the veans binary against this workspace, returning stdout,
// stderr, and exit code.
func (h *Harness) Run(t *testing.T, ws *Workspace, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), h.Binary, args...)
	cmd.Dir = ws.Dir
	// Filter VEANS_* out of the inherited env before applying our
	// overrides — a developer's VEANS_TOKEN would otherwise mask the
	// per-test bot token via the env backend.
	cmd.Env = append(filterEnv(os.Environ(), "VEANS_"), envSlice(ws.envOverrides)...)
	var so, se bytes.Buffer
	cmd.Stdout = &so
	cmd.Stderr = &se
	err := cmd.Run()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return so.String(), se.String(), ee.ExitCode()
		}
		t.Fatalf("run veans %v: %v", args, err)
	}
	return so.String(), se.String(), 0
}

// CreateProject creates a fresh project owned by the admin user and returns
// it. Tests use a unique title to keep results isolated across parallel runs.
func (h *Harness) CreateProject(t *testing.T, title, identifier string) *client.Project {
	t.Helper()
	out, err := h.AdminClient.CreateProject(t.Context(),
		&client.Project{Title: title, Identifier: identifier})
	if err != nil {
		t.Fatalf("create project %q: %v", title, err)
	}
	return out
}

// FindKanbanView returns the first Kanban view of the project (Vikunja
// auto-creates one).
func (h *Harness) FindKanbanView(t *testing.T, projectID int64) *client.ProjectView {
	t.Helper()
	views, err := h.AdminClient.ListProjectViews(t.Context(), projectID)
	if err != nil {
		t.Fatalf("list views: %v", err)
	}
	for _, v := range views {
		if v.ViewKind == client.ViewKindKanban {
			return v
		}
	}
	t.Fatalf("no Kanban view on project %d", projectID)
	return nil
}

// GetTask fetches a task by ID for verification.
func (h *Harness) GetTask(t *testing.T, id int64) *client.Task {
	t.Helper()
	task, err := h.AdminClient.GetTask(t.Context(), id)
	if err != nil {
		t.Fatalf("get task %d: %v", id, err)
	}
	return task
}

func buildOrLocate() (string, error) {
	if env := os.Getenv("VEANS_BINARY"); env != "" {
		if abs, err := filepath.Abs(env); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}
	tmp, err := os.MkdirTemp("", "veans-bin-*")
	if err != nil {
		return "", err
	}
	bin := filepath.Join(tmp, "veans")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "./cmd/veans")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build veans: %w (output: %s)", err, out)
	}
	return bin, nil
}

func repoRoot() string {
	// e2e/helpers.go lives at <repo>/veans/e2e/helpers.go, so go up two.
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func envSlice(overrides map[string]string) []string {
	out := make([]string, 0, len(overrides))
	for k, v := range overrides {
		out = append(out, k+"="+v)
	}
	return out
}

// filterEnv returns env entries whose keys do NOT start with prefix.
func filterEnv(env []string, prefix string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, prefix) {
			out = append(out, kv)
		}
	}
	return out
}
