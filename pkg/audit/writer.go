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

package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/google/uuid"
)

var (
	mu           sync.Mutex
	initialized  bool
	logFile      *os.File
	logfilePath  string
	currentSize  int64
	maxSizeBytes int64
	maxAge       time.Duration
	lastSync     time.Time
)

// Init opens the audit log file.
// Safe to call again to re-read the config (used by tests).
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	closeLocked()

	logfilePath = config.AuditLogfile.GetString()
	if logfilePath == "" {
		logfilePath = filepath.Join(config.LogPath.GetString(), "audit.log")
	}
	maxSizeBytes = config.AuditRotationMaxSizeMB.GetInt64() * 1024 * 1024
	maxAge = time.Duration(config.AuditRotationMaxAge.GetInt64()) * 24 * time.Hour

	if err := os.MkdirAll(filepath.Dir(logfilePath), 0750); err != nil {
		return fmt.Errorf("could not create audit log directory: %w", err)
	}
	if err := openLogFileLocked(); err != nil {
		return err
	}

	initialized = true
	return nil
}

// Close closes the audit log file. Used by tests.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	closeLocked()
}

func closeLocked() {
	if logFile != nil {
		_ = logFile.Sync()
		_ = logFile.Close()
		logFile = nil
	}
	initialized = false
}

func openLogFileLocked() error {
	f, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not open audit log file %s: %w", logfilePath, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("could not stat audit log file %s: %w", logfilePath, err)
	}
	logFile = f
	currentSize = info.Size()
	return nil
}

// WriteAuditEvent writes one entry to the local audit log. A failed write is
// returned so the event router retries it.
func WriteAuditEvent(entry *Entry) error {
	if entry.EventID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("could not generate audit event id: %w", err)
		}
		entry.EventID = id.String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.Outcome == "" {
		entry.Outcome = OutcomeSuccess
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("could not marshal audit entry: %w", err)
	}

	mu.Lock()
	if !initialized {
		mu.Unlock()
		return fmt.Errorf("audit log not initialized")
	}

	if err := rotateIfNeededLocked(int64(len(line)) + 1); err != nil {
		mu.Unlock()
		return err
	}

	// A failed rotation can leave us without an open file — retry the open
	// here so writes self-heal via the router's retries instead of panicking.
	if logFile == nil {
		if err := openLogFileLocked(); err != nil {
			mu.Unlock()
			return err
		}
	}

	written, err := logFile.Write(append(line, '\n'))
	currentSize += int64(written)
	if err == nil && time.Since(lastSync) > time.Second {
		err = logFile.Sync()
		lastSync = time.Now()
	}
	mu.Unlock()

	if err != nil {
		return fmt.Errorf("could not write audit entry: %w", err)
	}

	return nil
}

func rotateIfNeededLocked(addition int64) error {
	if maxSizeBytes <= 0 || currentSize+addition <= maxSizeBytes {
		return nil
	}

	_ = logFile.Sync()
	_ = logFile.Close()
	logFile = nil

	rotatedPath := rotatedFileName(logfilePath, time.Now().UTC())
	if err := os.Rename(logfilePath, rotatedPath); err != nil {
		// Reopen the original so logging continues even if rotation failed.
		if openErr := openLogFileLocked(); openErr != nil {
			return errors.Join(fmt.Errorf("could not rotate audit log: %w", err), openErr)
		}
		return fmt.Errorf("could not rotate audit log: %w", err)
	}

	cleanupRotatedFiles()

	return openLogFileLocked()
}

func rotatedFileName(path string, now time.Time) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext) + "-" + now.Format("20060102T150405.000") + ext
}

func cleanupRotatedFiles() {
	if maxAge <= 0 {
		return
	}

	ext := filepath.Ext(logfilePath)
	pattern := strings.TrimSuffix(logfilePath, ext) + "-*" + ext
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Errorf("Could not list rotated audit log files: %s", err)
		return
	}

	cutoff := time.Now().Add(-maxAge)
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		if err := os.Remove(match); err != nil {
			log.Errorf("Could not remove old audit log file %s: %s", match, err)
		}
	}
}
