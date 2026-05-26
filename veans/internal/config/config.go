// Package config reads and writes the per-repo .veans.yml file. The schema
// pins the project, view, canonical buckets, and bot identity so subsequent
// veans calls have everything they need without round-tripping to the server.
package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"code.vikunja.io/veans/internal/output"
)

// Filename is the canonical config name. Walked upward from cwd by Find.
const Filename = ".veans.yml"

// Config is the on-disk shape of .veans.yml.
type Config struct {
	Server            string  `yaml:"server"`
	ProjectID         int64   `yaml:"project_id"`
	ProjectIdentifier string  `yaml:"project_identifier,omitempty"`
	ViewID            int64   `yaml:"view_id"`
	Buckets           Buckets `yaml:"buckets"`
	Bot               Bot     `yaml:"bot"`

	path string `yaml:"-"`
}

// Buckets maps the five canonical statuses to bucket IDs.
type Buckets struct {
	Todo       int64 `yaml:"todo"`
	InProgress int64 `yaml:"in_progress"`
	InReview   int64 `yaml:"in_review"`
	Done       int64 `yaml:"done"`
	Scrapped   int64 `yaml:"scrapped"`
}

// Bot identifies the Vikunja bot user veans operates as.
type Bot struct {
	Username string `yaml:"username"`
	UserID   int64  `yaml:"user_id"`
}

// Path returns the absolute path the config was loaded from (or written to).
func (c *Config) Path() string { return c.path }

// FormatTaskID renders a numeric task index in the project's preferred form:
// PROJ-NN if the project has an identifier, #NN otherwise.
func (c *Config) FormatTaskID(index int64) string {
	if c.ProjectIdentifier != "" {
		return fmt.Sprintf("%s-%d", c.ProjectIdentifier, index)
	}
	return fmt.Sprintf("#%d", index)
}

// Find walks upward from cwd looking for .veans.yml. Returns ErrNotFound if
// none is reachable.
func Find(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	dir := start
	for {
		candidate := filepath.Join(dir, Filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// ErrNotFound is returned by Find when no .veans.yml is reachable.
var ErrNotFound = errors.New(".veans.yml not found in any parent directory")

// Load reads .veans.yml from `path`. Use Find to locate it first.
func Load(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, output.Wrap(output.CodeNotConfigured, err, "no .veans.yml at %s", path)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(buf, &c); err != nil {
		return nil, output.Wrap(output.CodeValidation, err, "parse %s: %v", path, err)
	}
	c.path = path
	return &c, nil
}

// Save writes the config to its path (must be set on the struct).
func (c *Config) Save() error {
	if c.path == "" {
		return errors.New("Save: no path set on Config")
	}
	buf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path, buf, 0o644)
}

// SaveAs writes the config to a specific path (and updates c.path).
func (c *Config) SaveAs(path string) error {
	c.path = path
	return c.Save()
}

// RepoRoot returns the root of the git repo containing `start` (defaulting
// to cwd). When `start` is not in a git repo, RepoRoot returns the absolute
// `start` so callers can still derive a sensible bot username.
func RepoRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = start
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	abs, _ := filepath.Abs(start)
	return abs, nil
}

// SuggestedBotUsername proposes `bot-<reponame>` from a repo root path.
// Vikunja's username validator allows lowercase, digits, hyphens — we fold
// the basename to a safe shape.
func SuggestedBotUsername(root string) string {
	base := filepath.Base(root)
	var b strings.Builder
	b.WriteString("bot-")
	for _, r := range strings.ToLower(base) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == ' ' || r == '.':
			b.WriteRune('-')
		default:
			// drop other characters silently
		}
	}
	// Collapse runs of hyphens.
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.TrimRight(out, "-")
}
