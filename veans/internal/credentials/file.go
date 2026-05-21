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
//
// Writes are atomic (tmp file + rename) and serialized by an advisory
// flock on a sibling .lock file so two concurrent `veans login` runs can
// each install their token without losing the other's.
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

// NewFileBackend builds a FileBackend rooted at `path`, or
// ~/.config/veans/credentials.yml when path is "".
func NewFileBackend(path string) *FileBackend {
	if path == "" {
		path = defaultCredsPath()
	}
	return &FileBackend{path: path}
}

func (b *FileBackend) Name() string { return "file" }
func (b *FileBackend) Path() string { return b.path }

// defaultCredsPath returns ~/.config/veans/credentials.yml, falling back to
// "" (which signals an error to NewFileBackend's caller) when there's no
// resolvable home directory. We deliberately do not honor XDG_CONFIG_HOME
// — it gave us a path-traversal seam for no real benefit, since the
// agent-only audience runs in a known environment.
func defaultCredsPath() string {
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return ""
	}
	return filepath.Join(h, ".config", "veans", "credentials.yml")
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

// save writes the schema atomically (tmpfile + rename) at mode 0600 and
// re-asserts the mode on the destination inode in case an earlier write
// left a wider mode behind.
func (b *FileBackend) save(s *fileSchema) (rerr error) {
	if err := os.MkdirAll(filepath.Dir(b.path), 0o700); err != nil {
		return err
	}
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(b.path), ".credentials-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// On any error before the rename completes, drop the half-written
	// temp file. tmpPath is the return of CreateTemp on a directory we
	// own, so gosec's path-traversal warning doesn't apply.
	defer func() {
		if rerr != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath) //nolint:gosec // G703: tmpPath came from os.CreateTemp on a dir we control
		}
	}()
	// CreateTemp opens at 0600 already, but be defensive: an inherited
	// umask shouldn't matter for CreateTemp on POSIX, but explicit is
	// cheaper than debugging later.
	if err := tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := tmp.Write(buf); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// gosec/G703: both paths come from FileBackend state (b.path is set
	// from defaultCredsPath or an explicit constructor arg, tmpPath from
	// CreateTemp on the same dir); neither is runtime user-influenceable.
	if err := os.Rename(tmpPath, b.path); err != nil { //nolint:gosec
		return err
	}
	// Belt-and-braces: a pre-existing destination at 0644 keeps its mode
	// across Rename on some filesystems. Narrow it.
	return os.Chmod(b.path, 0o600)
}

func (b *FileBackend) Get(server, account string) (string, error) {
	if b.path == "" {
		return "", ErrNotFound
	}
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

// Set serializes load → mutate → atomic save under an advisory flock on
// `<path>.lock` so two concurrent `veans login` runs don't clobber each
// other's tokens.
func (b *FileBackend) Set(server, account, token string) error {
	if b.path == "" {
		return errors.New("no credentials path: $HOME is unset and no explicit path was given")
	}
	if err := os.MkdirAll(filepath.Dir(b.path), 0o700); err != nil {
		return err
	}
	lockF, err := os.OpenFile(b.path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer lockF.Close()
	if err := flockExclusive(lockF); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer flockUnlock(lockF) //nolint:errcheck // unlock-on-close is sufficient

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
