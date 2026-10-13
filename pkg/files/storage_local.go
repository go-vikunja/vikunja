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

package files

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// localStorage implements FileStorage using the OS filesystem.
// All paths are resolved relative to basePath.
type localStorage struct {
	basePath string
}

func newLocalStorage(basePath string) *localStorage {
	return &localStorage{basePath: basePath}
}

func (l *localStorage) path(name string) string {
	return filepath.Join(l.basePath, name)
}

func (l *localStorage) Open(name string) (io.ReadCloser, error) {
	return os.Open(l.path(name))
}

func (l *localStorage) OpenWithSize(name string) (io.ReadCloser, int64, error) {
	f, err := os.Open(l.path(name))
	if err != nil {
		return nil, 0, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

// Write goes through a temporary file, so no reader sees a partial file and a failed write leaves none behind.
func (l *localStorage) Write(name string, content io.ReadSeeker, _ uint64) error {
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return err
	}

	target := l.path(name)
	tmp := target + ".tmp-" + rand.Text()
	// Not os.CreateTemp: its 0600 would change the mode of stored files.
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, content)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp, target)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}

func (l *localStorage) Stat(name string) (os.FileInfo, error) {
	return os.Stat(l.path(name))
}

func (l *localStorage) Remove(name string) error {
	return os.Remove(l.path(name))
}

func (l *localStorage) MkdirAll(p string, perm os.FileMode) error {
	return os.MkdirAll(l.path(p), perm)
}

func (l *localStorage) Ensure() error {
	// Anything other than "not there" is left to ValidateBasePath to report.
	if _, err := os.Stat(l.basePath); !errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err := os.MkdirAll(l.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create file storage directory at %s: %w%s", l.basePath, err, storageDiagSuffix(l.basePath))
	}

	return nil
}

func (l *localStorage) ValidateBasePath() error {
	info, err := os.Stat(l.basePath)
	if err != nil {
		return fmt.Errorf("failed to access file storage directory at %s: %w%s", l.basePath, err, storageDiagSuffix(l.basePath))
	}

	if !info.IsDir() {
		return fmt.Errorf("file storage path exists but is not a directory: %s", l.basePath)
	}

	return nil
}
