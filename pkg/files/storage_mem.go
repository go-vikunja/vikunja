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
	"bytes"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

// memStorage is an in-memory FileStorage for tests.
type memStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func newMemStorage() *memStorage {
	return &memStorage{files: make(map[string][]byte)}
}

func (m *memStorage) Open(name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.files[name]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memStorage) Write(name string, content io.ReadSeeker, _ uint64) error {
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return err
	}
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[name] = data
	return nil
}

func (m *memStorage) Stat(name string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.files[name]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
	}
	return &memFileInfo{
		name: path.Base(name),
		size: int64(len(data)),
	}, nil
}

func (m *memStorage) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.files[name]; !ok {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	delete(m.files, name)
	return nil
}

func (*memStorage) MkdirAll(string, os.FileMode) error {
	return nil
}

type memFileInfo struct {
	name string
	size int64
}

func (fi *memFileInfo) Name() string       { return fi.name }
func (fi *memFileInfo) Size() int64        { return fi.size }
func (fi *memFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *memFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *memFileInfo) IsDir() bool        { return false }
func (fi *memFileInfo) Sys() interface{}   { return nil }
