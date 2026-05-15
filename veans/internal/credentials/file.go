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

package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileBackend persists credentials to ~/.config/veans/credentials.yml at
// mode 0600. It's the fallback when no keychain is available (CI, Docker,
// headless servers) and is the implicit backend e2e tests use.
type FileBackend struct {
	path string
}

type fileEntry struct {
	Server  string `yaml:"server"`
	Account string `yaml:"account"`
	Token   string `yaml:"token"`
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
