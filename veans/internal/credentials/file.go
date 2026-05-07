package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FileBackend persists credentials to ~/.config/veans/credentials.yml at
// mode 0600. It's the fallback when no keychain is available (CI, Docker,
// headless servers) and is the implicit backend e2e tests use.
//
// The schema includes a `scope` field that's always empty in v0 but reserved
// for project-scoped tokens once Vikunja gains them — the same store can
// hold both kinds without migration.
type FileBackend struct {
	path string
}

type fileEntry struct {
	Server    string     `yaml:"server"`
	Account   string     `yaml:"account"`
	Scope     string     `yaml:"scope,omitempty"`
	Token     string     `yaml:"token"`
	ExpiresAt *time.Time `yaml:"expires_at,omitempty"`
}

type fileSchema struct {
	Credentials []fileEntry `yaml:"credentials"`
}

// NewFileBackend builds a FileBackend rooted at `path`, or the platform
// default (~/.config/veans/credentials.yml, honoring XDG_CONFIG_HOME) when
// path is "".
func NewFileBackend(path string) *FileBackend {
	if path == "" {
		path = defaultCredsPath()
	}
	return &FileBackend{path: path}
}

func (b *FileBackend) Name() string { return "file" }
func (b *FileBackend) Path() string { return b.path }

func defaultCredsPath() string {
	if c := os.Getenv("XDG_CONFIG_HOME"); c != "" {
		return filepath.Join(c, "veans", "credentials.yml")
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".config", "veans", "credentials.yml")
	}
	return filepath.Join(".", "credentials.yml")
}

func (b *FileBackend) load() (*fileSchema, error) {
	buf, err := os.ReadFile(b.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &fileSchema{}, nil
		}
		return nil, err
	}
	var s fileSchema
	if err := yaml.Unmarshal(buf, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", b.path, err)
	}
	return &s, nil
}

func (b *FileBackend) save(s *fileSchema) error {
	if err := os.MkdirAll(filepath.Dir(b.path), 0o700); err != nil {
		return err
	}
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(b.path, buf, 0o600)
}

func (b *FileBackend) Get(server, account string) (string, error) {
	s, err := b.load()
	if err != nil {
		return "", err
	}
	for _, e := range s.Credentials {
		if e.Server == server && e.Account == account {
			return e.Token, nil
		}
	}
	return "", ErrNotFound
}

func (b *FileBackend) Set(server, account, token string) error {
	s, err := b.load()
	if err != nil {
		return err
	}
	for i, e := range s.Credentials {
		if e.Server == server && e.Account == account {
			s.Credentials[i].Token = token
			return b.save(s)
		}
	}
	s.Credentials = append(s.Credentials, fileEntry{
		Server:  server,
		Account: account,
		Token:   token,
	})
	return b.save(s)
}

func (b *FileBackend) Delete(server, account string) error {
	s, err := b.load()
	if err != nil {
		return err
	}
	out := s.Credentials[:0]
	removed := false
	for _, e := range s.Credentials {
		if e.Server == server && e.Account == account {
			removed = true
			continue
		}
		out = append(out, e)
	}
	if !removed {
		return ErrNotFound
	}
	s.Credentials = out
	return b.save(s)
}
