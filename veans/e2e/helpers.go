// Package e2e is the integration suite for veans. It assumes a running
// Vikunja API at VEANS_E2E_API_URL and admin/seed credentials in
// VEANS_E2E_ADMIN_TOKEN (or VEANS_E2E_ADMIN_USER + VEANS_E2E_ADMIN_PASS).
//
// The suite never provisions Vikunja itself — locally, point it at a dev
// instance; in CI, the workflow spins one up the same way the frontend
// Playwright suite does.
package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/veans/internal/client"
)

// Harness bundles a built veans binary and an authenticated admin client
// for verifying side effects on the server. It also tracks the base env
// (HOME / XDG_CONFIG_HOME overrides) every runVeans invocation inherits.
type Harness struct {
	binary       string
	apiURL       string
	adminToken   string
	adminClient  *client.Client
	suiteStartTS time.Time
}

// SkipIfNotConfigured calls t.Skip if the suite hasn't been pointed at a
// Vikunja instance. Intended for the top of TestMain / TestXxx so plain
// `go test ./...` doesn't fail on contributors who haven't set up the env.
func SkipIfNotConfigured(t *testing.T) {
	t.Helper()
	if os.Getenv("VEANS_E2E_API_URL") == "" {
		t.Skip("VEANS_E2E_API_URL not set — skipping e2e")
	}
}

// New builds (or reuses) the veans binary, mints/loads an admin token via
// the env, and returns a Harness ready to drive tests.
func New(t *testing.T) *Harness {
	t.Helper()
	SkipIfNotConfigured(t)

	apiURL := strings.TrimRight(os.Getenv("VEANS_E2E_API_URL"), "/")
	binary, err := buildOrLocate()
	if err != nil {
		t.Fatalf("locate veans binary: %v", err)
	}

	tok := os.Getenv("VEANS_E2E_ADMIN_TOKEN")
	if tok == "" {
		user, pass := os.Getenv("VEANS_E2E_ADMIN_USER"), os.Getenv("VEANS_E2E_ADMIN_PASS")
		if user == "" || pass == "" {
			t.Fatal("set VEANS_E2E_ADMIN_TOKEN or VEANS_E2E_ADMIN_USER + VEANS_E2E_ADMIN_PASS")
		}
		c := client.New(apiURL, "")
		resp, err := c.Login(context.Background(), &client.LoginRequest{
			Username: user, Password: pass, LongToken: true,
		})
		if err != nil {
			t.Fatalf("admin login: %v", err)
		}
		tok = resp.Token
	}

	return &Harness{
		binary:       binary,
		apiURL:       apiURL,
		adminToken:   tok,
		adminClient:  client.New(apiURL, tok),
		suiteStartTS: time.Now(),
	}
}

// Workspace creates a per-test git repo in a TempDir, with HOME and
// XDG_CONFIG_HOME pointed at TempDirs so the credential store falls back
// to its file backend rather than touching the developer's keychain.
type Workspace struct {
	Dir          string
	Home         string
	XDGConfig    string
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
	xdg := t.TempDir()

	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "veans-e2e@example.com"},
		{"git", "config", "user.name", "veans-e2e"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
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
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(c, " "), err, out)
		}
	}

	return &Workspace{
		Dir:        dir,
		Home:       home,
		XDGConfig:  xdg,
		ConfigPath: filepath.Join(dir, ".veans.yml"),
		envOverrides: map[string]string{
			"HOME":            home,
			"XDG_CONFIG_HOME": xdg,
		},
	}
}

// Run executes the veans binary against this workspace, returning stdout,
// stderr, and exit code.
func (h *Harness) Run(t *testing.T, ws *Workspace, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(h.binary, args...)
	cmd.Dir = ws.Dir
	cmd.Env = append(os.Environ(), envSlice(ws.envOverrides)...)
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

// AdminClient returns the admin-authenticated client for verification.
func (h *Harness) AdminClient() *client.Client { return h.adminClient }

// AdminToken returns the admin's bearer token (handy for --token flows).
func (h *Harness) AdminToken() string { return h.adminToken }

// APIURL returns the configured Vikunja base URL.
func (h *Harness) APIURL() string { return h.apiURL }

// CreateProject creates a fresh project owned by the admin user and returns
// it. Tests use a unique title to keep results isolated across parallel runs.
func (h *Harness) CreateProject(t *testing.T, title, identifier string) *client.Project {
	t.Helper()
	body := map[string]any{"title": title}
	if identifier != "" {
		body["identifier"] = identifier
	}
	var out client.Project
	if err := h.adminClient.Do(context.Background(), "PUT", "/projects", nil, body, &out); err != nil {
		t.Fatalf("create project %q: %v", title, err)
	}
	return &out
}

// FindKanbanView returns the first Kanban view of the project (Vikunja
// auto-creates one).
func (h *Harness) FindKanbanView(t *testing.T, projectID int64) *client.ProjectView {
	t.Helper()
	views, err := h.adminClient.ListProjectViews(context.Background(), projectID)
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
	task, err := h.adminClient.GetTask(context.Background(), id)
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
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/veans")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build veans: %v\n%s", err, out)
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
