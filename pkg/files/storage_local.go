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

func (l *localStorage) Write(name string, content io.ReadSeeker, _ uint64) error {
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return err
	}

	f, err := os.Create(l.path(name))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, content)
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
