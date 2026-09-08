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

package migration

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.vikunja.io/api/pkg/log"
)

// spoolDirName holds queued imports below the OS temp dir. net/http already
// spools multipart uploads there, so this adds no new capacity requirement.
const spoolDirName = "vikunja-imports"

// errEmptySpoolName guards the zero value: filepath.Base("") is ".", which
// would resolve to the spool directory itself.
var errEmptySpoolName = errors.New("no spooled upload name given")

func spoolDir() (string, error) {
	dir := filepath.Join(os.TempDir(), spoolDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("could not create the import spool directory: %w", err)
	}
	return dir, nil
}

// spoolPath keeps a name from the event queue inside the spool directory.
func spoolPath(name string) (string, error) {
	if name == "" {
		return "", errEmptySpoolName
	}
	dir, err := spoolDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.Base(name)), nil
}

// SpoolUpload copies an uploaded import to disk so the background job can still
// read it after the request that carried it is gone. Only the returned name
// travels through the event queue.
func SpoolUpload(src io.Reader) (name string, size int64, err error) {
	dir, err := spoolDir()
	if err != nil {
		return "", 0, err
	}

	f, err := os.CreateTemp(dir, "import-*")
	if err != nil {
		return "", 0, fmt.Errorf("could not create the import spool file: %w", err)
	}
	defer f.Close()

	size, err = io.Copy(f, src)
	if err != nil {
		_ = os.Remove(f.Name())
		return "", 0, fmt.Errorf("could not spool the import file: %w", err)
	}

	return filepath.Base(f.Name()), size, nil
}

// OpenSpooledUpload opens an upload previously stored by SpoolUpload.
func OpenSpooledUpload(name string) (*os.File, error) {
	path, err := spoolPath(name)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

// RemoveSpooledUpload deletes an upload once its import is done. Failures are
// logged rather than returned: the import itself already succeeded or failed.
func RemoveSpooledUpload(name string) {
	path, err := spoolPath(name)
	if err != nil {
		log.Errorf("[Migration] Could not resolve the spooled import file %q: %s", name, err)
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Errorf("[Migration] Could not remove the spooled import file %q: %s", name, err)
	}
}

// CleanupSpooledUploads removes uploads left behind by an instance that died
// mid-import. The queue is in-process, so nothing can still be reading them.
func CleanupSpooledUploads() {
	dir, err := spoolDir()
	if err != nil {
		log.Errorf("[Migration] Could not clean up spooled imports: %s", err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Errorf("[Migration] Could not read the import spool directory: %s", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			log.Errorf("[Migration] Could not remove the orphaned import file %q: %s", entry.Name(), err)
		}
	}
}
